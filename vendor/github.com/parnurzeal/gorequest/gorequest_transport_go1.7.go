// +build go1.7
// +build !go1.8

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
        MaxIdleConns:           oldTransport.MaxIdleConns,
        MaxIdleConnsPerHost:    oldTransport.MaxIdleConnsPerHost,
        ResponseHeaderTimeout:  oldTransport.ResponseHeaderTimeout,
        ExpectContinueTimeout:  oldTransport.ExpectContinueTimeout,
        TLSNextProto:           oldTransport.TLSNextProto,
        // new in go1.7
        DialContext:            oldTransport.DialContext,
        IdleConnTimeout:        oldTransport.IdleConnTimeout,
        MaxResponseHeaderBytes: oldTransport.MaxResponseHeaderBytes,
    }
}
