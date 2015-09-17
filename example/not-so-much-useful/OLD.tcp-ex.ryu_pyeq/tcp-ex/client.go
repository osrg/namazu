package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)


type Worker struct {
	Name string
	Conn net.Conn
	Reader *bufio.Reader
	SendChan chan string
	QuitChan chan bool
}

func NewWorker(name string, tcpAddr *net.TCPAddr) *Worker {
	worker := &Worker{Name: name,
		Conn: nil,
		SendChan: make(chan string),
		QuitChan: make(chan bool)}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	worker.Conn = conn
	worker.Reader = bufio.NewReader(conn)
	go worker.routine()
	return worker
}

func (worker *Worker) routine() {
	for {
		select {
		case msg := <-worker.SendChan:
			fmt.Printf("%s: SEND: %s\n", worker.Name, msg)
			_, err := worker.Conn.Write([]byte(fmt.Sprintf("[REQ]worker=%s, msg=%s\r\n",worker.Name,msg)))
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}
			recvMsgChan := make(chan string)
			go func(){
				recvMsgAsBytes, _, err := worker.Reader.ReadLine()
				if err != nil {
					println(err.Error())
					os.Exit(1)
				}
				recvMsgChan <- string(recvMsgAsBytes[:])
			}()
			recvMsg := <- recvMsgChan
			fmt.Printf("%s: RECV: %s\n", worker.Name, recvMsg)
		case <-worker.QuitChan:
			fmt.Printf("%s: QUIT\n", worker.Name)
			worker.Conn.Close()
			return
		}
	}
}


func (worker *Worker) Send(msg string) {
	worker.SendChan <- msg
}

func (worker *Worker) Close() {
	worker.QuitChan <- true
}

func clientMain(servAddr string, nmessages int, nworkers int) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	workers := make([]Worker, nworkers)
	for i := 0; i < nworkers; i++ {
		workers[i] = *NewWorker(fmt.Sprintf("w%02d", i), tcpAddr)
		defer workers[i].Close()
	}

	for i := 0; i < nmessages; i++ {
		msg := fmt.Sprintf("msg%03d", i)
		worker := workers[i%nworkers]
		worker.Send(msg)
	}
}
