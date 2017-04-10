package main

import (
	"fmt"
	"io"
	"net"

	"github.com/golang/glog"
	"github.com/paultag/sniff/parser"
)

type server struct {
	Hostname string
	IP       string
	Port     int
}

type proxy struct {
	ServerList []*server
	Default    *server
}

func (p *proxy) Get(host string) *server {
	for _, s := range p.ServerList {
		if s.Hostname == host {
			return s
		}
	}

	return &server{
		Hostname: "localhost",
		IP:       "127.0.0.1",
		Port:     442,
	}
}

func (p *proxy) Handle(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 4096)

	length, err := conn.Read(data)
	if err != nil {
		glog.V(4).Infof("error reading the first 4k of the connection: %s", err)
		return
	}

	var proxy *server
	hostname, err := parser.GetHostname(data[:])
	if err == nil {
		glog.V(2).Infof("parsed hostname: %s", hostname)
		proxy = p.Get(hostname)
		if proxy == nil {
			return
		}
	} else {
		proxy = p.Default
		if proxy == nil {
			return
		}
	}

	clientConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", proxy.IP, proxy.Port))
	if err != nil {
		return
	}
	defer clientConn.Close()

	_, err = clientConn.Write(data[:length])
	if err != nil {
		clientConn.Close()
	}
	pipe(clientConn, conn)
}

func pipe(client, server net.Conn) {
	doCopy := func(s, c net.Conn, cancel chan<- bool) {
		io.Copy(s, c)
		cancel <- true
	}

	cancel := make(chan bool, 2)

	go doCopy(server, client, cancel)
	go doCopy(client, server, cancel)

	select {
	case <-cancel:
		return
	}
}
