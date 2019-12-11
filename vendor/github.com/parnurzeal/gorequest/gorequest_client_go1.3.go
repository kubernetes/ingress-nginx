// +build go1.3

package gorequest

import (
    "time"
    "net/http"
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
    s.Client.Timeout = oldClient.Timeout
    s.Client.CheckRedirect = oldClient.CheckRedirect
}


func (s *SuperAgent) Timeout(timeout time.Duration) *SuperAgent {
    s.safeModifyHttpClient()
    s.Client.Timeout = timeout
    return s
}