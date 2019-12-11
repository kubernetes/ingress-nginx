// +build go1.2
// +build !go1.3

package gorequest

import (
    "net"
    "net/http"
    "time"
)


// we don't want to mess up other clones when we modify the client..
// so unfortantely we need to create a new client
func (s *SuperAgent) safeModifyHttpClient() {
    if !s.isClone {
        return
    }
    oldClient := s.Client
    s.Client = &http.Client{}
    s.Client.Jar = oldClient.Jar
    s.Client.Transport = oldClient.Transport
    s.Client.CheckRedirect = oldClient.CheckRedirect
}

// I'm not sure how this func will work with Clone.
func (s *SuperAgent) Timeout(timeout time.Duration) *SuperAgent {
    s.Transport.Dial = func(network, addr string) (net.Conn, error) {
        conn, err := net.DialTimeout(network, addr, timeout)
        if err != nil {
            s.Errors = append(s.Errors, err)
            return nil, err
        }
        conn.SetDeadline(time.Now().Add(timeout))
        return conn, nil
    }
    return s
}