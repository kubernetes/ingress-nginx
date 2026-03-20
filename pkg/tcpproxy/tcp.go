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

package tcpproxy

import (
	"fmt"
	"io"
	"net"

	"k8s.io/klog/v2"

	"pault.ag/go/sniff/parser"
)

// TCPServer describes a server that works in passthrough mode.
type TCPServer struct {
	Hostname      string
	IP            string
	Port          int
	ProxyProtocol bool
}

// TCPProxy describes the passthrough servers and a default as catch all.
type TCPProxy struct {
	ServerList []*TCPServer
	Default    *TCPServer
}

// Get returns the TCPServer to use for a given host.
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
	// It appears that the ClientHello must fit into *one* TLSPlaintext message:
	// When a client first connects to a server, it is REQUIRED to send the ClientHello as its first TLS message.
	// Source: https://datatracker.ietf.org/doc/html/rfc8446#section-4.1.2
	//
	// length:  The length (in bytes) of the following TLSPlaintext.fragment. The length MUST NOT exceed 2^14 bytes.
	// An endpoint that receives a record that exceeds this length MUST terminate the connection with a "record_overflow" alert.
	// Source: https://datatracker.ietf.org/doc/html/rfc8446#section-5.1
	// bytes 0  : content type
	// bytes 1-2: legacy version
	// bytes 3-4: length
	// bytes 5+ : message
	// https://en.wikipedia.org/wiki/Transport_Layer_Security#TLS_record
	// Thus, we need to allocate 5 + 16384 bytes
	data := make([]byte, parser.TLSHeaderLength+16384)

	// read the tls header first
	_, err := io.ReadFull(conn, data[:parser.TLSHeaderLength])
	if err != nil {
		klog.V(4).ErrorS(err, "Error reading TLS header from the connection")
		return
	}
	// get the total data length then read the rest
	length := min(int(data[3])<<8+int(data[4])+parser.TLSHeaderLength, len(data))
	_, err = io.ReadFull(conn, data[parser.TLSHeaderLength:length])
	if err != nil {
		klog.V(4).ErrorS(err, "Error reading data from the connection")
		return
	}

	proxy := p.Default
	hostname, err := parser.GetHostname(data)
	if err == nil {
		klog.V(4).InfoS("TLS Client Hello", "host", hostname)
		proxy = p.Get(hostname)
	}

	if proxy == nil {
		klog.V(4).InfoS("There is no configured proxy for SSL connections.")
		return
	}

	hostPort := net.JoinHostPort(proxy.IP, fmt.Sprintf("%v", proxy.Port))
	klog.V(4).InfoS("passing to", "hostport", hostPort)
	clientConn, err := net.Dial("tcp", hostPort)
	if err != nil {
		klog.V(4).ErrorS(err, "error dialing proxy", "ip", proxy.IP, "port", proxy.Port, "hostname", proxy.Hostname)
		return
	}
	defer clientConn.Close()

	if proxy.ProxyProtocol {
		// write out the Proxy Protocol header
		localAddr, ok := conn.LocalAddr().(*net.TCPAddr)
		if !ok {
			klog.Errorf("unexpected type: %T", conn.LocalAddr())
		}
		remoteAddr, ok := conn.RemoteAddr().(*net.TCPAddr)
		if !ok {
			klog.Errorf("unexpected type: %T", conn.RemoteAddr())
		}
		protocol := "UNKNOWN"
		if remoteAddr.IP.To4() != nil {
			protocol = "TCP4"
		} else if remoteAddr.IP.To16() != nil {
			protocol = "TCP6"
		}
		proxyProtocolHeader := fmt.Sprintf("PROXY %s %s %s %d %d\r\n", protocol, remoteAddr.IP.String(), localAddr.IP.String(), remoteAddr.Port, localAddr.Port)
		klog.V(4).InfoS("Writing Proxy Protocol", "header", proxyProtocolHeader)
		_, err = fmt.Fprint(clientConn, proxyProtocolHeader)
	}
	if err != nil {
		klog.ErrorS(err, "Error writing Proxy Protocol header")
		clientConn.Close()
	} else {
		_, err = clientConn.Write(data[:length])
		if err != nil {
			klog.Errorf("Error writing the first %d bytes of proxy data: %v", length, err)
			clientConn.Close()
		}
	}

	pipe(clientConn, conn)
}

func pipe(client, server net.Conn) {
	doCopy := func(s, c net.Conn, cancel chan<- bool) {
		//nolint:errcheck // No need to catch these errors
		io.Copy(s, c)
		cancel <- true
	}

	cancel := make(chan bool, 2)

	go doCopy(server, client, cancel)
	go doCopy(client, server, cancel)

	<-cancel
}
