// Copyright 2021 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet v4 monitoring client package
package teomon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
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

type Peers []*Metric

type Metric struct {
	Address    string
	AppName    string
	AppShort   string
	AppVersion string
	Online     bool

	bslice.ByteSlice
}

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
