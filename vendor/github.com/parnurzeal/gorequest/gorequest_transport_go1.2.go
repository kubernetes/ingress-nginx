// +build go1.2
// +build !go1.3

package gorequest

import (
    "net/http"
)

// does a shallow clone of the transport
func (s *SuperAgent) safeModifyTransport() {
    if !s.isClone {
        return
    }
    oldTransport := s.Transport
    s.Transport = &http.Transport{
        Proxy:                  oldTransport.Proxy,
        Dial:                   oldTransport.Dial,
        TLSClientConfig:        oldTransport.TLSClientConfig,
        DisableKeepAlives:      oldTransport.DisableKeepAlives,
        DisableCompression:     oldTransport.DisableCompression,
        MaxIdleConnsPerHost:    oldTransport.MaxIdleConnsPerHost,
        ResponseHeaderTimeout:  oldTransport.ResponseHeaderTimeout,
    }
}
