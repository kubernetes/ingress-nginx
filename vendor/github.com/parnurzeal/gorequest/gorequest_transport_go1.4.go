// +build go1.4
// +build !go1.6

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
        TLSHandshakeTimeout:    oldTransport.TLSHandshakeTimeout,
        DisableKeepAlives:      oldTransport.DisableKeepAlives,
        DisableCompression:     oldTransport.DisableCompression,
        MaxIdleConnsPerHost:    oldTransport.MaxIdleConnsPerHost,
        ResponseHeaderTimeout:  oldTransport.ResponseHeaderTimeout,
        // new in go1.4
        DialTLS:                oldTransport.DialTLS,
    }
}
