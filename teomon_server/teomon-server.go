// Copyright 2021 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet v4 monitoring server package
package teomon_server

import (
	"time"

	"github.com/kirill-scherba/teomon/teomon"
	"github.com/kirill-scherba/teonet"
)

// New teonet monitoring
func New(teo *teonet.Teonet, appName, appShort, appLong, appVersion string) (mon *Teomon) {
	mon = new(Teomon)
	mon.API = teo.NewAPI(appName, appShort, appLong, appVersion)
	teo.AddReader(mon.Commands().Reader())
	mon.CheckOnline()
	return
}

type Teomon struct {
	*teonet.API
	peers teomon.Peers
}

// Commands teonet monitoring service API commands
func (teo *Teomon) Commands() *Teomon {
	teo.Add(
		// Command Hello (test)
		func() teonet.APInterface {
			var cmdApi *teonet.APIData
			cmdApi = teonet.MakeAPI2().
				SetCmd(teo.Cmd(129)).                 // Command number cmd = 129
				SetName("hello").                     // Command name
				SetShort("get 'hello name' message"). // Short description
				SetUsage("<name string>").            // Usage (input parameter)
				SetReturn("<answer string>").         // Return (output parameters)
				// Command reader (execute when command received)
				SetReader(func(c *teonet.Channel, p *teonet.Packet, data []byte) bool {
					data = append([]byte("Hello "), data...)
					teo.SendAnswer(cmdApi, c, data, p)
					return true
				}).
				SetAnswerMode( /* teonet.CmdAnswer | */ teonet.DataAnswer)
			return cmdApi
		}(),

		// Command Metric. Application send metric to monitor
		teonet.MakeAPI2().
			SetCmd(teo.Cmd(teomon.CmdMetric)).  // Command number cmd = 130
			SetName("metric").                  // Command name
			SetShort("send metric to monitor"). // Short description
			SetUsage("<metric MonitorMetric>"). // Usage (input parameter)
			// Command reader (execute when command received)
			SetReader(func(c *teonet.Channel, p *teonet.Packet, data []byte) bool {
				teo.Log().Println("got metric command from", c)
				metric := teomon.NewMetric()
				err := metric.UnmarshalBinary(data)
				if err != nil {
					teo.Log().Println("unmarshal metric error:", err)
					return true
				}
				metric.Address = c.Address()
				metric.Params.Add(teomon.ParamOnline, true)
				teo.peers.Add(metric)
				return true
			}).SetAnswerMode(teonet.NoAnswer),

		// Command Parameter. Application send parameter to monitor
		teonet.MakeAPI2().
			SetCmd(teo.Cmd(teomon.CmdParameter)).     // Command number cmd = 131
			SetName("parameter").                     // Command name
			SetShort("send parameter to monitor").    // Short description
			SetUsage("<parameter MonitorParameter>"). // Usage (input parameter)
			// Command reader (execute when command received)
			SetReader(func(c *teonet.Channel, p *teonet.Packet, data []byte) bool {
				teo.Log().Println("got parameter command from", c)
				param := teomon.NewParameter()
				err := param.UnmarshalBinary(data)
				if err != nil {
					teo.Log().Println("unmarshal parameter error:", err)
					return true
				}
				metric, ok := teo.peers.Get(c.Address())
				if !ok {
					teo.Log().Println("can't find peer with address:", c.Address())
					return true
				}
				metric.Params.Add(param.Name, param.Value)
				return true
			}).SetAnswerMode(teonet.NoAnswer),

		// Command List get list of peers
		func() teonet.APInterface {
			var cmdApi *teonet.APIData
			cmdApi = teonet.MakeAPI2().
				SetCmd(teo.Cmd(teo.CmdNext())). // Command number cmd = 132
				SetName("list").                // Command name
				SetShort("get list of peers").  // Short description
				// SetUsage("<name string>").      // Usage (input parameter)
				SetReturn("<answer []*Metric>"). // Return (output parameters)
				// Command reader (execute when command received)
				SetReader(func(c *teonet.Channel, p *teonet.Packet, data []byte) bool {
					teo.Log().Println("got list command from", c)
					out, err := teo.peers.MarshalBinary()
					if err != nil {
						teo.Log().Println("marshal to binary error:", err)
						return true
					}
					teo.SendAnswer(cmdApi, c, out, p)
					return true
				}).
				SetAnswerMode( /* teonet.CmdAnswer | */ teonet.DataAnswer)
			return cmdApi
		}(),
	)
	return teo
}

// CheckOnline check peers connected now and set online parameter
func (teo *Teomon) CheckOnline() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			teo.peers.Each(func(m *teomon.Metric) {
				online := teo.Connected(m.Address)
				m.Params.Add(teomon.ParamOnline, online)
			})
		}
	}()
}
