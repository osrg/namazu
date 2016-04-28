package main

import (
	"fmt"
	"os"

	"github.com/osrg/namazu/nmz/cli"
	"github.com/osrg/namazu/nmz/explorepolicy"
	"github.com/osrg/namazu/nmz/historystorage"
	"github.com/osrg/namazu/nmz/signal"
	"github.com/osrg/namazu/nmz/util/config"
)

// implements ExplorePolicy interface
type MyPolicy struct {
	nextActionChan chan signal.Action
}

func NewMyPolicy() explorepolicy.ExplorePolicy {
	return &MyPolicy{
		nextActionChan: make(chan signal.Action),
	}
}

// implements ExplorePolicy interface
func (p *MyPolicy) Name() string {
	return "mypolicy"
}

// implements ExplorePolicy interface
func (p *MyPolicy) LoadConfig(cfg config.Config) error {
	return nil
}

// implements ExplorePolicy interface
func (p *MyPolicy) SetHistoryStorage(storage historystorage.HistoryStorage) error {
	return nil
}

// implements ExplorePolicy interface
func (p *MyPolicy) ActionChan() chan signal.Action {
	return p.nextActionChan
}

// implements ExplorePolicy interface
func (p *MyPolicy) QueueEvent(event signal.Event) {
	// Possible events:
	//  - JavaFunctionEvent
	//  - PacketEvent
	//  - FilesystemEvent
	//  - ProcSetEvent (Linux procfs)
	//  - LogEvent
	fmt.Printf("Event: %s\n", event)
	// You can also inject fault actions
	//  - PacketFaultAction
	//  - FilesystemFaultAction
	//  - ProcSetSchedAction
	//  - ShellAction
	action, err := event.DefaultAction()
	if err != nil {
		panic(err)
	}
	// send in a goroutine so as to make the function non-blocking.
	// (Note that namazu/util/queue/TimeBoundedQueue provides
	// better semantics and determinism, this is just an example.)
	go func() {
		fmt.Printf("Action ready: %s\n", action)
		p.nextActionChan <- action
		fmt.Printf("Action passed: %s\n", action)
	}()
}

func main() {
	fmt.Println("Hello Namazu + mypolicy")

	explorepolicy.RegisterPolicy("mypolicy", NewMyPolicy)

	os.Exit(cli.CLIMain(os.Args))
}
