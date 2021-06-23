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
)

const (
	appName    = "Teonet monitoring server application"
	appShort   = "teomon"
	appVersion = "0.2.4"
	appLong    = ""
)

func main() {
	teonet.Logo(appName, appVersion)

	// Command line arguments
	var params struct {
		appShort  string
		logLevel  string
		showTrudp bool
		port      int
	}
	flag.StringVar(&params.appShort, "app-short", appShort, "application short name")
	flag.BoolVar(&params.showTrudp, "u", false, "show trudp statistic")
	flag.IntVar(&params.port, "p", 0, "local port")
	flag.StringVar(&params.logLevel, "log-level", "NONE", "log level")
	flag.Parse()

	// Initial Teonet
	teo, err := teonet.New(params.appShort, params.port,
		params.showTrudp, params.logLevel, teonet.Log(),
	)
	if err != nil {
		teo.Log().Println("can't init Teonet, error:", err)
		return
	}

	// Connect to teonet
	for teo.Connect() != nil {
		time.Sleep(1 * time.Second)
	}

	// Start teonet monitor server
	teomon_server.New(teo, appName, appShort, appLong, appVersion)

	// Teonet address
	fmt.Printf("Teonet addres: %s\n\n", teo.Address())

	// sleep forever
	select {}
}
