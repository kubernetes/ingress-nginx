package addr

import (
	"net"
)

// Suggest suggests a address a process can listen on. It returns
// a tuple consisting of a free port and the hostname resolved to its IP.
func Suggest() (port int, resolvedHost string, err error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}
	port = l.Addr().(*net.TCPAddr).Port
	defer func() {
		err = l.Close()
	}()
	resolvedHost = addr.IP.String()
	return
}
