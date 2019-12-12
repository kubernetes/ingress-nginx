// +build go1.8

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
        DialContext:            oldTransport.DialContext,
        Dial:                   oldTransport.Dial,
        DialTLS:                oldTransport.DialTLS,
        TLSClientConfig:        oldTransport.TLSClientConfig,
        TLSHandshakeTimeout:    oldTransport.TLSHandshakeTimeout,
        DisableKeepAlives:      oldTransport.DisableKeepAlives,
        DisableCompression:     oldTransport.DisableCompression,
        MaxIdleConns:           oldTransport.MaxIdleConns,
        MaxIdleConnsPerHost:    oldTransport.MaxIdleConnsPerHost,
        IdleConnTimeout:        oldTransport.IdleConnTimeout,
        ResponseHeaderTimeout:  oldTransport.ResponseHeaderTimeout,
        ExpectContinueTimeout:  oldTransport.ExpectContinueTimeout,
        TLSNextProto:           oldTransport.TLSNextProto,
        MaxResponseHeaderBytes: oldTransport.MaxResponseHeaderBytes,
        // new in go1.8
        ProxyConnectHeader:     oldTransport.ProxyConnectHeader,
    }
}
