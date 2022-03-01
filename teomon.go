// Copyright 2021 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet v4 monitoring client package
package teomon

import (
	"bytes"
	"encoding/binary"
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

const (
	CmdMetric    byte = 130
	CmdParameter byte = 131
)

type TeonetInterface interface {
	WhenConnectedDisconnected(f func())
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
		data, _ := m.MarshalBinary()
		data = append([]byte{CmdMetric}, data...)
		teo.SendTo(address, data)
		mon.SendParam(ParamPeers, teocheck.NumPeers())
	})

	// Connect to monitor
	for teo.ConnectTo(address) != nil {
		time.Sleep(1 * time.Second)
	}

	// Process connected/disconnected events and send Parameter "peers" to monitor
	teocheck.WhenConnectedDisconnected(func() {
		mon.SendParam(ParamPeers, teocheck.NumPeers())
	})

	return
}

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

type Metric struct {
	Address      string
	AppName      string
	AppShort     string
	AppVersion   string
	TeoVersion   string
	AppStartTime time.Time
	New          bool

	Params *Parameters

	bslice.ByteSlice
}

const (
	ParamOnline = "online"
	ParamPeers  = "peers"
)

func NewMetric() (m *Metric) {
	m = new(Metric)
	m.NewParams()
	return
}

func (m *Metric) NewParams() {
	m.Params = &Parameters{m: make(map[string]interface{})}
}

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

	if err = binary.Write(buf, binary.LittleEndian, uint16(len(m.Params.m))); err != nil {
		return
	}
	m.Params.RLock()
	defer m.Params.RUnlock()
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

type Parameters struct {
	m map[string]interface{}
	sync.RWMutex
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

func NewParameter() (p *Parameter) {
	p = new(Parameter)
	return
}

type Parameter struct {
	Name  string
	Value interface{}

	bslice.ByteSlice
}

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

type Peers []*Metric

func (p Peers) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)

	l := uint16(len(p))
	binary.Write(buf, binary.LittleEndian, l)
	for _, m := range p {
		d, _ := m.MarshalBinary()
		m.WriteSlice(buf, d)
	}

	data = buf.Bytes()
	return
}

func (p *Peers) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)

	*p = nil
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
		*p = append(*p, m)
	}

	return
}

// Save peers to file
func (p Peers) Save(file string) (err error) {

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
func (p Peers) find(address string) (m *Metric, idx int, ok bool) {
	for idx, m = range p {
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
		(*p)[i] = metric
		return
	}

	// Add new
	metric.Params.m = make(map[string]interface{})
	metric.Params.Add(ParamOnline, true)
	*p = append(*p, metric)
}

// Get peer metric by address
func (p Peers) Get(address string) (m *Metric, ok bool) {
	m, _, ok = p.find(address)
	return
}

// Each execute callback for each Metric
func (p Peers) Each(f func(m *Metric)) {
	for _, m := range p {
		f(m)
	}
}

func (p Peers) String() (str string) {
	sort.Slice(p, func(i, j int) bool { return p[i].AppShort < p[j].AppShort })

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

	for _, m := range p {
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

	for i, m := range p {
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
			if name == ParamOnline || name == ParamPeers {
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