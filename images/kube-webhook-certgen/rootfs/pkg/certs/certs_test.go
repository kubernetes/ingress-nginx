package certs

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func handler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello World")
}

func TestCertificateCreation(t *testing.T) {
	ca, cert, key := GenerateCerts("localhost")

	c, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    caCertPool,
			ServerName: "localhost",
		},
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	ts.TLS = &tls.Config{Certificates: []tls.Certificate{c}}
	ts.StartTLS()
	defer ts.Close()

	client := &http.Client{Transport: tr}
	res, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Response code was %v; want 200", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte("Hello World")

	if bytes.Compare(expected, body) != 0 {
		t.Errorf("Response body was '%v'; want '%v'", expected, body)
	}
}
