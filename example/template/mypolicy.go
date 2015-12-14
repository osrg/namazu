package main

import (
	"fmt"
	"github.com/osrg/earthquake/earthquake/cli"
	"github.com/osrg/earthquake/earthquake/explorepolicy"
	"github.com/osrg/earthquake/earthquake/historystorage"
	"github.com/osrg/earthquake/earthquake/signal"
	"os"
)

// implements ExplorePolicy interface
type MyPolicy struct {
	nextActionChan chan signal.Action
}

// implements ExplorePolicy interface
func (p *MyPolicy) Init(storage historystorage.HistoryStorage, param map[string]interface{}) {
	fmt.Printf("Initializing, param=%#v\n", param)
}

// implements ExplorePolicy interface
func (p *MyPolicy) Name() string {
	return "mypolicy"
}

// implements ExplorePolicy interface
func (p *MyPolicy) GetNextActionChan() chan signal.Action {
	return p.nextActionChan
}

// implements ExplorePolicy interface
func (p *MyPolicy) QueueNextEvent(event signal.Event) {
	// Possible events:
	//  - JavaFunctionEvent
	//  - PacketEvent
	//  - FilesystemEvent
	//  - LogEvent
	fmt.Printf("Event: %s\n", event)
	// You can also inject fault actions
	//  - PacketFaultAction
	//  - FilesystemFaultAction
	//  - ShellAction
	action, err := event.DefaultAction()
	if err != nil {
		panic(err)
	}
	// send in a goroutine so as to make the function non-blocking.
	// (Note that earthquake/util/queue/TimeBoundedQueue provides
	// better semantics and determinism, this is just an example.)
	go func() {
		fmt.Printf("Action ready: %s\n", action)
		p.nextActionChan <- action
		fmt.Printf("Action passed: %s\n", action)
	}()
}

func NewMyPolicy() explorepolicy.ExplorePolicy {
	return &MyPolicy{
		nextActionChan: make(chan signal.Action),
	}
}

func main() {
	fmt.Println("Hello Earthquake + mypolicy")

	explorepolicy.RegisterPolicy("mypolicy", NewMyPolicy)

	os.Exit(cli.CLIMain(os.Args))
}
