// +build go1.6
// +build !go1.7

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
        DialTLS:                oldTransport.DialTLS,
        TLSClientConfig:        oldTransport.TLSClientConfig,
        TLSHandshakeTimeout:    oldTransport.TLSHandshakeTimeout,
        DisableKeepAlives:      oldTransport.DisableKeepAlives,
        DisableCompression:     oldTransport.DisableCompression,
        MaxIdleConnsPerHost:    oldTransport.MaxIdleConnsPerHost,
        ResponseHeaderTimeout:  oldTransport.ResponseHeaderTimeout,
        // new in 1.6
        ExpectContinueTimeout:  oldTransport.ExpectContinueTimeout,
        TLSNextProto:           oldTransport.TLSNextProto,
    }
}
