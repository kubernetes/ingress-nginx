package main

import (
	"fmt"
	"io"
	"net"

	"github.com/golang/glog"
	"github.com/paultag/sniff/parser"
)

type server struct {
	Hostname      string
	IP            string
	Port          int
	ProxyProtocol bool
}

type proxy struct {
	ServerList []*server
	Default    *server
}

func (p *proxy) Get(host string) *server {
	if p.ServerList == nil {
		return p.Default
	}

	for _, s := range p.ServerList {
		if s.Hostname == host {
			return s
		}
	}

	return p.Default
}

func (p *proxy) Handle(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 4096)

	length, err := conn.Read(data)
	if err != nil {
		glog.V(4).Infof("error reading the first 4k of the connection: %s", err)
		return
	}

	proxy := p.Default
	hostname, err := parser.GetHostname(data[:])
	if err == nil {
		glog.V(4).Infof("parsed hostname from TLS Client Hello: %s", hostname)
		proxy = p.Get(hostname)
	}

	if proxy == nil {
		glog.V(4).Infof("there is no configured proxy for SSL connections")
		return
	}

	clientConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", proxy.IP, proxy.Port))
	if err != nil {
		return
	}
	defer clientConn.Close()

	if proxy.ProxyProtocol {
		//Write out the proxy-protocol header
		localAddr := conn.LocalAddr().(*net.TCPAddr)
		remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
		protocol := "UNKNOWN"
		if remoteAddr.IP.To4() != nil {
			protocol = "TCP4"
		} else if remoteAddr.IP.To16() != nil {
			protocol = "TCP6"
		}
		proxyProtocolHeader := fmt.Sprintf("PROXY %s %s %s %d %d\r\n", protocol, remoteAddr.IP.String(), localAddr.IP.String(), remoteAddr.Port, localAddr.Port)
		glog.V(4).Infof("Writing proxy protocol header - %s", proxyProtocolHeader)
		_, err = fmt.Fprintf(clientConn, proxyProtocolHeader)
	}
	if err != nil {
		glog.Errorf("unexpected error writing proxy-protocol header: %s", err)
		clientConn.Close()
	} else {
		_, err = clientConn.Write(data[:length])
		if err != nil {
			glog.Errorf("unexpected error writing first 4k of proxy data: %s", err)
			clientConn.Close()
		}
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
