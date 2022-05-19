// Copyright 2021-22 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet v5 monitoring client package
package teomon

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kirill-scherba/bslice"
)

// Comand constant
const (
	CmdMetric    byte = 130
	CmdParameter byte = 131

	version = "0.5.11"
)

// TeonetInterface define teonet functions used in teomon
type TeonetInterface interface {
	WhenConnectedDisconnected(f func(e byte))
	WhenConnectedTo(address string, f func())
	ConnectTo(address string, attr ...interface{}) error
	SendTo(address string, data []byte, attr ...interface{}) (int, error)
	Address() string
	NumPeers() int
}

// Connect to monitor peer and send metric
func Connect(teo TeonetInterface, address string, m Metric, t ...TeonetInterface) (mon *Monitor) {

	mon = new(Monitor)
	mon.teo = teo
	mon.address = address

	// Which teonet check for connected: the same or from t parameter
	var teocheck = teo
	if len(t) > 0 {
		teocheck = t[0]
	}

	// When connected to monitor
	teo.WhenConnectedTo(address, func() {
		m.NewParams()

		// Send metric
		data, _ := m.MarshalBinary()
		data = append([]byte{CmdMetric}, data...)
		teo.SendTo(address, data)

		// Send parameter 'number of peers'
		mon.SendParam(ParamPeers, teocheck.NumPeers())

		// Send parameter 'host name'
		if h, err := os.Hostname(); err == nil {
			mon.SendParam(ParamHost, h)
		}

		// Send parameter 'machineid'
		if id, err := getMachineID(); err == nil {
			mon.SendParam(ParamMachineID, id)
		}
	})

	// Connect to monitor
	for teo.ConnectTo(address) != nil {
		time.Sleep(1 * time.Second)
	}

	// Process connected/disconnected events and send Parameter "peers" to monitor
	teocheck.WhenConnectedDisconnected(func(e byte) {
		numPeers := teocheck.NumPeers()
		if e == 5 /* EventDisconnected */ {
			numPeers--
		}
		mon.SendParam(ParamPeers, numPeers)
	})

	return
}

// Teonet monitor struct
type Monitor struct {
	teo     TeonetInterface
	address string
}

// SendParam send parameter to monitor
func (mon Monitor) SendParam(name string, value interface{}) {
	p := NewParameter()
	p.Name = name
	p.Value = value
	data, _ := p.MarshalBinary()
	data = append([]byte{CmdParameter}, data...)
	mon.teo.SendTo(mon.address, data)
}

// Metric contain metric struct and methods receiver
type Metric struct {
	Address      string
	AppName      string
	AppShort     string
	AppVersion   string
	TeoVersion   string
	AppStartTime time.Time
	New          bool
	Params       *Parameters
	bslice.ByteSlice
}

// Param constant
const (
	ParamOnline    = "online"
	ParamPeers     = "peers"
	ParamHost      = "host"
	ParamMachineID = "machineid"
	MayOffline     = "mayoffline"
)

// NewMetric create new metric object
func NewMetric() (m *Metric) {
	m = new(Metric)
	m.NewParams()
	return
}

// MarshalBinary binary marshal Metric struct
func (m Metric) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)

	m.WriteSlice(buf, []byte(m.Address))
	m.WriteSlice(buf, []byte(m.AppName))
	m.WriteSlice(buf, []byte(m.AppShort))
	m.WriteSlice(buf, []byte(m.AppVersion))
	m.WriteSlice(buf, []byte(m.TeoVersion))
	//
	d, err := m.AppStartTime.MarshalBinary()
	if err != nil {
		return
	}
	m.WriteSlice(buf, d)
	//
	binary.Write(buf, binary.LittleEndian, m.New)

	m.Params.RLock()
	defer m.Params.RUnlock()

	if err = binary.Write(buf, binary.LittleEndian, uint16(len(m.Params.m))); err != nil {
		return
	}
	for name, val := range m.Params.m {
		p := Parameter{Name: name, Value: val}
		data, err := p.MarshalBinary()
		if err != nil {
			return nil, err
		}
		m.WriteSlice(buf, data)
	}

	data = buf.Bytes()
	return
}

// UnmarshalBinary binary unmarshal Metric struct
func (m *Metric) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)

	if m.Address, err = m.ReadString(buf); err != nil {
		return
	}
	if m.AppName, err = m.ReadString(buf); err != nil {
		return
	}
	if m.AppShort, err = m.ReadString(buf); err != nil {
		return
	}
	if m.AppVersion, err = m.ReadString(buf); err != nil {
		return
	}
	if m.TeoVersion, err = m.ReadString(buf); err != nil {
		return
	}

	d, err := m.ReadSlice(buf)
	if err != nil {
		return
	}
	if err = m.AppStartTime.UnmarshalBinary(d); err != nil {
		return
	}

	if err = binary.Read(buf, binary.LittleEndian, &m.New); err != nil {
		return
	}

	var l uint16
	if err = binary.Read(buf, binary.LittleEndian, &l); err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		p := Parameter{}
		data, err := m.ReadSlice(buf)
		if err != nil {
			return err
		}
		p.UnmarshalBinary(data)
		m.Params.Add(p.Name, p.Value)
	}

	return
}

// Parameters is metric parameters struct and methods receiver
type Parameters struct {
	m map[string]interface{}
	sync.RWMutex
}

// NewParams create new params object
func (m *Metric) NewParams() {
	m.Params = &Parameters{m: make(map[string]interface{})}
}

// add or update parameter
func (p *Parameters) Add(name string, val interface{}) {
	p.Lock()
	defer p.Unlock()
	p.m[name] = val
}

// get parameter
func (p *Parameters) Get(name string) (val interface{}, ok bool) {
	p.RLock()
	defer p.RUnlock()
	val, ok = p.m[name]
	return
}

// Each execute callback for each Parameter
func (p *Parameters) Each(f func(name string, value interface{})) {
	p.RLock()
	defer p.RUnlock()
	for n, v := range p.m {
		f(n, v)
	}
}

// Parameter struct and methods receiver
type Parameter struct {
	Name  string
	Value interface{}
	bslice.ByteSlice
}

// NewParameter create new parameter
func NewParameter() (p *Parameter) {
	p = new(Parameter)
	return
}

// MarshalBinary binary marshal Parameter struct
func (p Parameter) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)

	p.WriteSlice(buf, []byte(p.Name))
	t := reflect.TypeOf(p.Value).String()
	p.WriteSlice(buf, []byte(t))
	switch t {
	case "string":
		p.WriteSlice(buf, []byte(p.Value.(string)))
	case "[]uint8":
		p.WriteSlice(buf, p.Value.([]byte))
	case "int":
		binary.Write(buf, binary.LittleEndian, int32(p.Value.(int)))
	default:
		binary.Write(buf, binary.LittleEndian, p.Value)
	}

	data = buf.Bytes()
	return
}

// UnmarshalBinary binary unmarshal Parameter struct
func (p *Parameter) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)

	if p.Name, err = p.ReadString(buf); err != nil {
		return
	}
	var t string
	if t, err = p.ReadString(buf); err != nil {
		return
	}
	// fmt.Println("type:", t)

	switch t {
	case "bool":
		var val bool
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.Value = val

	case "int":
		var val int32
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.Value = int(val)

	case "int32":
		var val int32
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.Value = val

	case "uint32":
		var val uint32
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.Value = val

	case "float64":
		var val float64
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.Value = val

	case "string":
		var val string
		if val, err = p.ReadString(buf); err != nil {
			return
		}
		p.Value = val

	case "[]uint8":
		var val []byte
		if val, err = p.ReadSlice(buf); err != nil {
			return
		}
		p.Value = val

	default:
		err = fmt.Errorf("unmarshal error - unsupported type: %s", t)
	}

	return
}

// Peers struct and methods receiver
type Peers struct {
	metrics []*Metric
	*sync.RWMutex
}

// NewPeers create new Peers struct
func NewPeers() (p *Peers) {
	p = new(Peers)
	p.RWMutex = new(sync.RWMutex)
	return
}

// MarshalBinary binary marshal Peers struct
func (p *Peers) MarshalBinary() (data []byte, err error) {
	p.RLock()
	defer p.RUnlock()

	buf := new(bytes.Buffer)
	l := uint16(len(p.metrics))
	binary.Write(buf, binary.LittleEndian, l)
	for _, m := range p.metrics {
		d, _ := m.MarshalBinary()
		m.WriteSlice(buf, d)
	}
	data = buf.Bytes()
	return
}

// UnmarshalBinary binary unmarshal Peers struct
func (p *Peers) UnmarshalBinary(data []byte) (err error) {
	p.Lock()
	defer p.Unlock()

	buf := bytes.NewBuffer(data)
	p.metrics = nil
	var l uint16
	if err = binary.Read(buf, binary.LittleEndian, &l); err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		m := NewMetric()
		var d []byte
		d, err = m.ReadSlice(buf)
		if err != nil {
			return
		}
		err = m.UnmarshalBinary(d)
		if err != nil {
			return
		}
		p.metrics = append(p.metrics, m)
	}
	return
}

// Save peers to file
func (p *Peers) Save(file string) (err error) {

	f, err := os.Create(file)
	if err != nil {
		return
	}

	// Set all metrics New value to false
	p.Each(func(m *Metric) {
		m.New = false
	})

	data, err := p.MarshalBinary()
	if err != nil {
		return
	}

	_, err = f.Write(data)

	return
}

// Load peers from file
func (p *Peers) Load(file string) (err error) {

	const bufferSize = 1024 * 1024

	// Open file
	f, err := os.Open(file)
	if err != nil {
		return
	}

	// Read file data
	data := make([]byte, bufferSize)
	n, err := f.Read(data)
	if err != nil {
		return
	}
	if n == bufferSize {
		err = errors.New("too small read buffer")
		return
	}

	// Unmarshal config data
	err = p.UnmarshalBinary(data[:n])
	if err != nil {
		return
	}

	return
}

// find metric by address
func (p *Peers) find(address string, unsafe ...bool) (m *Metric, idx int, ok bool) {
	if len(unsafe) == 0 || !unsafe[0] {
		p.RLock()
		defer p.RUnlock()
	}

	for idx, m = range p.metrics {
		if m.Address == address {
			ok = true
			return
		}
	}
	return
}

// Add or Update metric
func (p *Peers) Add(metric *Metric) {

	// Update if exists
	if _, i, ok := p.find(metric.Address); ok {
		p.Lock()
		defer p.Unlock()

		p.metrics[i] = metric
		return
	}

	// Add new
	metric.Params.m = make(map[string]interface{})
	metric.Params.Add(ParamOnline, true)

	p.Lock()
	defer p.Unlock()
	p.metrics = append(p.metrics, metric)
}

// Get peer metric by address
func (p *Peers) Get(address string) (m *Metric, ok bool) {
	m, _, ok = p.find(address)
	return
}

// Del peer by address
func (p *Peers) Del(address string) (m *Metric, ok bool) {
	p.Lock()
	defer p.Unlock()

	m, idx, ok := p.find(address, true)
	if !ok {
		return
	}

	switch {
	case idx == 0:
		p.metrics = p.metrics[1:]
	case idx == len(p.metrics)-1:
		p.metrics = p.metrics[:len(p.metrics)-1]
	default:
		p.metrics = append(p.metrics[:idx], p.metrics[idx+1:]...)
	}

	return
}

// Each execute callback for each Metric
func (p *Peers) Each(f func(m *Metric)) {
	p.RLock()
	defer p.RUnlock()

	for _, m := range p.metrics {
		f(m)
	}
}

// sortMetrics sort metrics with Online (offline first) and AppShort
func (p Peers) sortMetrics(metrics []*Metric) {
	sort.Slice(metrics, func(i, j int) bool {
		online1, _ := metrics[i].Params.Get(ParamOnline)
		online2, _ := metrics[j].Params.Get(ParamOnline)

		// If online parameter has valid type bool sort by online
		if reflect.TypeOf(online1).Kind() == reflect.Bool && reflect.TypeOf(online2).Kind() == reflect.Bool {
			onl1 := online1.(bool)
			onl2 := online2.(bool)
			switch {
			case !onl1 && onl2:
				return true
			case onl1 && !onl2:
				return false
			}
		}

		return metrics[i].AppShort < metrics[j].AppShort
	})
}

// String return string which contain Peers table
func (p Peers) String() (str string) {

	// Calculate max columns len
	var l struct {
		appShort   int
		appVersion int
		teoVersion int
		address    int
		online     int
		peers      int
		start      int
	}

	timeFormat := "2006-01-02 15:04:05"

	// Sort metrics
	p.sortMetrics(p.metrics)

	for _, m := range p.metrics {
		if len := len(m.AppShort); len > l.appShort {
			l.appShort = len
		}
		if len := len(m.AppVersion); len > l.appVersion {
			l.appVersion = len
		}
		if len := len(m.TeoVersion); len > l.teoVersion {
			l.teoVersion = len
		}
		if len := len(m.Address); len > l.address {
			l.address = len
		}
		start := fmt.Sprint(m.AppStartTime.Format(timeFormat))
		if len := len(start); len > l.start {
			l.start = len
		}
	}
	l.online = 6
	l.peers = 5

	numFields := reflect.TypeOf(l).NumField()

	line := strings.Repeat("-",
		l.appShort+l.appVersion+l.teoVersion+l.address+l.online+l.peers+l.start+
			5+4+(numFields-1)*3+2,
	) + "\n"

	str += line
	str += fmt.Sprintf("  # | %-*s | n | %-*s | %-*s | %-*s | online | peers | start time \n",
		l.appShort, "name", l.appVersion, "ver", l.teoVersion, "teo", l.address, "address")
	str += line

	for i, m := range p.metrics {
		online, _ := m.Params.Get(ParamOnline)
		peers, _ := m.Params.Get(ParamPeers)
		start := fmt.Sprint(m.AppStartTime.Format(timeFormat))
		newPeer := "n"
		if !m.New {
			newPeer = " "
		}
		str += fmt.Sprintf(" %2d | %-*s | %s | %-*s | %-*s | %-*s | %-*v | %*v | %*s \n",
			i+1,
			l.appShort, m.AppShort,
			newPeer,
			l.appVersion, m.AppVersion,
			l.teoVersion, m.TeoVersion,
			l.address, m.Address,
			l.online, online,
			l.peers, peers,
			l.start, start,
		)
		var numParams = 0
		m.Params.Each(func(name string, value interface{}) {
			switch name {
			case ParamOnline, ParamPeers, ParamHost, ParamMachineID, MayOffline:
				return
			}
			str += fmt.Sprintf("   %s: %v\n", name, value)
			numParams++
		})
		if numParams > 0 {
			str += "\n"
		}
	}
	str += line[:len(line)-1]

	return
}

// Json return string which contain Peers in json format
func (p Peers) Json() (data []byte, err error) {
	p.RLock()
	defer p.RUnlock()

	// Sort metrics
	p.sortMetrics(p.metrics)

	type Pmetric struct {
		Metric
		Online    interface{}
		Peers     interface{}
		Host      interface{}
		MachineID interface{}
	}

	var pmetrics []Pmetric

	// Add common parameters to output json
	for _, m := range p.metrics {
		mayoffline, _ := m.Params.Get(MayOffline)
		online, _ := m.Params.Get(ParamOnline)
		peers, _ := m.Params.Get(ParamPeers)
		host, _ := m.Params.Get(ParamHost)
		id, _ := m.Params.Get(ParamMachineID)
		pm := Pmetric{
			Metric:     *m,
			MayOffline: mayoffline,
			Online:     online,
			Peers:      peers,
			Host:       host,
			MachineID:  id,
		}
		pmetrics = append(pmetrics, pm)
	}

	// Marshal json
	return json.Marshal(pmetrics)
}
