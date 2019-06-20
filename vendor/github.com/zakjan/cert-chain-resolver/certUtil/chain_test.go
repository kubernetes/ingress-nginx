package certUtil

import (
	"crypto/x509"
    "testing"
	"os"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
)

func TestFetchCertificateChain(t *testing.T) {
	file, err := os.Open("../tests/comodo.der.crt")
	if err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}

	cert, err := x509.ParseCertificate(data)
	if err != nil {
		t.Error(err)
	}

	certs, err := FetchCertificateChain(cert)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 3, len(certs))
}
