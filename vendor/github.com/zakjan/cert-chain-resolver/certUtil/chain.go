package certUtil

import (
	"crypto/x509"
	"io/ioutil"
	"net/http"
)

func isSelfSigned(cert *x509.Certificate) bool {
	return cert.CheckSignatureFrom(cert) == nil
}

func isChainRootNode(cert *x509.Certificate) bool {
	if isSelfSigned(cert) {
		return true
	}
	return false
}

func FetchCertificateChain(cert *x509.Certificate) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	certs = append(certs, cert)

	for certs[len(certs)-1].IssuingCertificateURL != nil {
		parentURL := certs[len(certs)-1].IssuingCertificateURL[0]

		resp, err := http.Get(parentURL)
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		cert, err := DecodeCertificate(data)
		if err != nil {
			return nil, err
		}

		if isChainRootNode(cert) {
			break
		}

		certs = append(certs, cert)
	}

	return certs, nil
}

func AddRootCA(certs []*x509.Certificate) ([]*x509.Certificate, error) {
	lastCert := certs[len(certs)-1]

	chains, err := lastCert.Verify(x509.VerifyOptions{})
	if err != nil {
		if _, e := err.(x509.UnknownAuthorityError); e {
			return certs, nil
		}
		return nil, err
	}

	for _, cert := range chains[0] {
		if lastCert.Equal(cert) {
			continue
		}
		certs = append(certs, cert)
	}

	return certs, nil
}
