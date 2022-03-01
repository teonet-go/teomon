// Copyright 2021 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet v4 monitoring server application
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/kirill-scherba/teomon/teomon_server"
	"github.com/kirill-scherba/teonet"
	"github.com/kirill-scherba/tru/teolog"
)

const (
	appName    = "Teonet monitoring server application"
	appShort   = "teomon"
	appVersion = "0.3.0"
	appLong    = ""
)

var log = teolog.New()
var appStartTime = time.Now()

func main() {
	teonet.Logo(appName, appVersion)

	// Command line arguments
	var p struct {
		appShort  string
		loglevel  string
		logfilter string
		stat      bool
		hotkey    bool
		port      int
	}
	flag.StringVar(&p.appShort, "app-short", appShort, "application short name")
	flag.IntVar(&p.port, "p", 0, "local port")
	flag.BoolVar(&p.stat, "stat", false, "show statistic")
	flag.BoolVar(&p.hotkey, "hotkey", false, "start hotkey menu")
	flag.StringVar(&p.loglevel, "loglevel", "NONE", "set log level")
	flag.StringVar(&p.logfilter, "logfilter", "", "set log filter")
	flag.Parse()

	// Initial Teonet
	teo, err := teonet.New(p.appShort, p.port, teonet.ShowStat(p.stat),
		teonet.StartHotkey(p.hotkey), log, p.loglevel,
		teonet.Logfilter(p.logfilter),
	)
	if err != nil {
		panic("can't init Teonet, error: " + err.Error())
	}

	// Start teonet monitor server
	teomon_server.New(teo, appName, appShort, appLong, appVersion, appStartTime)

	// Connect to teonet
	for teo.Connect() != nil {
		time.Sleep(1 * time.Second)
	}

	// Teonet address
	fmt.Printf("Teonet addres: %s\n\n", teo.Address())

	// sleep forever
	select {}
}
