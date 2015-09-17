package main

import (
	"net"
	"fmt"
	"bufio"
	"time"
)

type Service struct {
	Conn net.Conn
	Reader *bufio.Reader
}

func NewService (conn net.Conn) *Service  {
	service := &Service{}
	service.Conn = conn
	service.Reader = bufio.NewReader(conn)
	return service
}

func (service *Service) processMessage(message string) {
	s := fmt.Sprintf ( "[RES]%s: %s: %s", 
		time.Now().Format(time.StampMilli), 
		service.Conn.RemoteAddr().String(), 
		message )
	fmt.Printf("%s\n", s)
	service.Conn.Write([]byte(fmt.Sprintf("%s\r\n", s)))
}

func (service *Service) getMessage() (string, error){
	var message string
	messageAsBytes, _, e := service.Reader.ReadLine()
	if e == nil {
		message = string(messageAsBytes[:])
	}
	return message, e
}

func (service *Service) Routine(){
	for {
		message, e := service.getMessage()
		if e != nil {
			service.Conn.Close()
			return
		}
		service.processMessage(message)
	}
}

func serverMain(listenTCPPort int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", listenTCPPort))
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	for {
		conn, _ := ln.Accept()
		service := NewService(conn)
                go service.Routine()
	}
}
