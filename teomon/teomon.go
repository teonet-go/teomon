// Copyright 2021 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet v4 monitoring client package
package teomon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kirill-scherba/bslice"
)

const (
	CmdMetric byte = 130
)

type TeonetInterface interface {
	WhenConnectedTo(address string, f func())
	ConnectTo(address string, attr ...interface{}) error
	SendTo(address string, data []byte, attr ...interface{}) (uint32, error)
}

// Connect to monitor peer and send metric
func Connect(teo TeonetInterface, address string, m Metric) {

	teo.WhenConnectedTo(address, func() {
		data, _ := m.MarshalBinary()
		data = append([]byte{CmdMetric}, data...)
		teo.SendTo(address, data)
	})

	for teo.ConnectTo(address) != nil {
		time.Sleep(1 * time.Second)
	}
}

type Metric struct {
	Address    string
	AppName    string
	AppShort   string
	AppVersion string

	Params *Parameters

	bslice.ByteSlice
}

const OnlineParam = "online"

func (m Metric) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)

	m.WriteSlice(buf, []byte(m.Address))
	m.WriteSlice(buf, []byte(m.AppName))
	m.WriteSlice(buf, []byte(m.AppShort))
	m.WriteSlice(buf, []byte(m.AppVersion))

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

	return
}

type Parameters struct {
	m map[string]interface{}
	sync.RWMutex
}

// add or update parameter
func (p *Parameters) Add(addr string, val interface{}) {
	p.Lock()
	defer p.Unlock()
	p.m[addr] = val
}

// get parameter
func (p *Parameters) Get(addr string) (val interface{}, ok bool) {
	p.RLock()
	defer p.RUnlock()
	val, ok = p.m[addr]
	return
}

type Parameter struct {
	address string
	name    string
	val     interface{}
	// t       reflect.Type

	bslice.ByteSlice
}

func (p Parameter) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)

	p.WriteSlice(buf, []byte(p.address))
	p.WriteSlice(buf, []byte(p.name))
	t := reflect.TypeOf(p.val).String()
	p.WriteSlice(buf, []byte(t))
	switch t {
	case "string":
		p.WriteSlice(buf, []byte(p.val.(string)))
	case "[]uint8":
		p.WriteSlice(buf, p.val.([]byte))
	case "int":
		binary.Write(buf, binary.LittleEndian, int32(p.val.(int)))
	default:
		binary.Write(buf, binary.LittleEndian, p.val)
	}

	data = buf.Bytes()
	return
}

func (p *Parameter) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)

	if p.address, err = p.ReadString(buf); err != nil {
		return
	}
	if p.name, err = p.ReadString(buf); err != nil {
		return
	}
	var t string
	if t, err = p.ReadString(buf); err != nil {
		return
	}
	fmt.Println("type:", t)

	switch t {
	case "bool":
		var val bool
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.val = val

	case "int":
		var val int32
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.val = int(val)

	case "int32":
		var val int32
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.val = val

	case "uint32":
		var val uint32
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.val = val

	case "float64":
		var val float64
		if err = binary.Read(buf, binary.LittleEndian, &val); err != nil {
			return
		}
		p.val = val

	case "string":
		var val string
		if val, err = p.ReadString(buf); err != nil {
			return
		}
		p.val = val

	case "[]uint8":
		var val []byte
		if val, err = p.ReadSlice(buf); err != nil {
			return
		}
		p.val = val

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
		var m Metric
		var d []byte
		d, err = m.ReadSlice(buf)
		if err != nil {
			return
		}
		err = m.UnmarshalBinary(d)
		if err != nil {
			return
		}
		*p = append(*p, &m)
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
	metric.Params.Add(OnlineParam, true)
	*p = append(*p, metric)
}

// List of metrics to data
func (p Peers) List() (data []byte) {
	data, _ = p.MarshalBinary()
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
		address    int
		online     int
	}
	for _, m := range p {
		if len := len(m.AppShort); len > l.appShort {
			l.appShort = len
		}
		if len := len(m.AppVersion); len > l.appVersion {
			l.appVersion = len
		}
		if len := len(m.Address); len > l.address {
			l.address = len
		}
	}
	l.online = 6

	line := strings.Repeat("-", l.appShort+l.appVersion+l.address+l.online+(4-1)*3+2) + "\n"

	str += line
	str += fmt.Sprintf(" %-*s | %-*s | %-*s | online\n",
		l.appShort, "name", l.appVersion, "ver", l.address, "address")
	str += line

	for _, m := range p {
		str += fmt.Sprintf(" %-*s | %-*s | %-*s |\n",
			l.appShort, m.AppShort,
			l.appVersion, m.AppVersion,
			l.address, m.Address,
		)
	}
	str += line[:len(line)-1]

	return
}
