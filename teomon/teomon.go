// Copyright 2021 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet v4 monitoring client package
package teomon

import (
	"bytes"
	"encoding/binary"
	"time"
)

const (
	CmdMetric byte = 130
)

type TeonetInterface interface {
	ConnectTo(address string, attr ...interface{}) error
	SendTo(address string, data []byte, attr ...interface{}) (uint32, error)
}

// Connect to monitor peer and send metric
func Connect(teo TeonetInterface, address string, m Metric) {

	for teo.ConnectTo(address) != nil {
		time.Sleep(1 * time.Second)
	}

	teo.SendTo(address, []byte{130})
}

type Peers []*Metric

type Metric struct {
	Address    string
	AppName    string
	AppShort   string
	AppVersion string

	ByteSlice
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

// find metric by address
func (p Peers) find(address string) (m *Metric, idx int, ok bool) {
	for idx, m = range p {
		if m.Address == address {
			ok = true
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

func (p Peers) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)

	l := uint16(len(p))
	binary.Write(buf, binary.LittleEndian, l)
	for i, m := range p {
		m.WriteSlice(buf, []byte(p[i].Address))
	}

	data = buf.Bytes()
	return
}

type ByteSlice struct{}

func (b ByteSlice) WriteSlice(buf *bytes.Buffer, data []byte) (err error) {
	if err = binary.Write(buf, binary.LittleEndian, uint16(len(data))); err != nil {
		return
	}
	err = binary.Write(buf, binary.LittleEndian, data)
	return
}

func (b ByteSlice) ReadSlice(buf *bytes.Buffer) (data []byte, err error) {
	var l uint16
	if err = binary.Read(buf, binary.LittleEndian, &l); err != nil {
		return
	}
	data = make([]byte, l)
	err = binary.Read(buf, binary.LittleEndian, data)
	return
}

func (b ByteSlice) ReadString(buf *bytes.Buffer) (data string, err error) {
	d, err := b.ReadSlice(buf)
	if err != nil {
		return
	}
	data = string(d)
	return
}
