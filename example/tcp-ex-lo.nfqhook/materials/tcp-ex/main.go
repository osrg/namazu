// forked from: https://github.com/osrg/earthquake/blob/master/earthquake/main.go
//
// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
)

type clientFlags struct {
	Peer string
	Messages int
	Workers int
}

type serverFlags struct {
	ListenTCPPort int
}

var (
	globalFlagset = flag.NewFlagSet("tcp-ex", flag.ExitOnError)
	globalFlags   = struct {
		LaunchClient bool
		LaunchServer bool
	}{}
	clientFlagset = flag.NewFlagSet("client", flag.ExitOnError)
	_clientFlags  = clientFlags{}

	serverFlagset = flag.NewFlagSet("server", flag.ExitOnError)
	_serverFlags  = serverFlags{}
)

func init() {
	globalFlagset.BoolVar(&globalFlags.LaunchClient, "client", false, "client mode")
	globalFlagset.BoolVar(&globalFlags.LaunchServer, "server", false, "server mode")

	clientFlagset.StringVar(&_clientFlags.Peer, "peer", "localhost:9999", "peer addr (host:port)")
	clientFlagset.IntVar(&_clientFlags.Messages, "messages", 16, "number of the all messages to send")
	clientFlagset.IntVar(&_clientFlags.Workers, "workers", 4, "number of the workers")
	
	serverFlagset.IntVar(&_serverFlags.ListenTCPPort, "listen-tcp-port", 9999, "TCP Port to listen on")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("specify -client or -server\n")
		os.Exit(1)
	}
	globalArgs := []string{os.Args[1]}

	globalFlagset.Parse(globalArgs)

	if globalFlags.LaunchClient && globalFlags.LaunchServer {
		fmt.Printf("don't specify both of -client and -server\n")
		os.Exit(1)
	}

	if globalFlags.LaunchClient {
		clientFlagset.Parse(os.Args[2:])
		clientMain(_clientFlags.Peer, _clientFlags.Messages, _clientFlags.Workers)
		return
	}

	if globalFlags.LaunchServer {
		serverFlagset.Parse(os.Args[2:])
		serverMain(_serverFlags.ListenTCPPort)
		return
	}

	fmt.Printf("THIS SHOULD NOT HAPPEN\n")
	os.Exit(1)
}
