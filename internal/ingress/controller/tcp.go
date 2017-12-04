/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"io"
	"net"

	"github.com/golang/glog"

	"github.com/paultag/sniff/parser"
)

// TCPServer describes a server that works in passthrough mode
type TCPServer struct {
	Hostname      string
	IP            string
	Port          int
	ProxyProtocol bool
}

// TCPProxy describes the passthrough servers and a default as catch all
type TCPProxy struct {
	ServerList []*TCPServer
	Default    *TCPServer
}

// Get returns the TCPServer to use
func (p *TCPProxy) Get(host string) *TCPServer {
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

// Handle reads enough information from the connection to extract the hostname
// and open a connection to the passthrough server.
func (p *TCPProxy) Handle(conn net.Conn) {
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
