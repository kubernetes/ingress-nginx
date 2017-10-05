package certUtil

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/fullsailor/pkcs7"
)

var pemStart = []byte("-----BEGIN ")
var certBlockType = "CERTIFICATE"

func IsPEM(data []byte) bool {
	return bytes.HasPrefix(data, pemStart)
}

func DecodeCertificates(data []byte) ([]*x509.Certificate, error) {
	if IsPEM(data) {
		var certs []*x509.Certificate

		for len(data) > 0 {
			var block *pem.Block

			block, data = pem.Decode(data)
			if block == nil {
				return nil, errors.New("Invalid certificate.")
			}
			if block.Type != certBlockType {
				return nil, errors.New("Invalid certificate.")
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, errors.New("Invalid certificate.")
			}

			certs = append(certs, cert)
		}

		return certs, nil
	} else {
		certs, err := x509.ParseCertificates(data)
		if err != nil {
			return nil, errors.New("Invalid certificate.")
		}

		return certs, nil
	}
}

func DecodeCertificate(data []byte) (*x509.Certificate, error) {
	if IsPEM(data) {
		block, _ := pem.Decode(data)
		if block == nil {
			return nil, errors.New("Invalid certificate.")
		}
		if block.Type != certBlockType {
			return nil, errors.New("Invalid certificate.")
		}

		data = block.Bytes
	}

	cert, err := x509.ParseCertificate(data)
	if err == nil {
		return cert, nil
	}

	p, err := pkcs7.Parse(data)
	if err == nil {
		return p.Certificates[0], nil
	}

	return nil, errors.New("Invalid certificate.")
}

func EncodeCertificate(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  certBlockType,
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

func EncodeCertificateDER(cert *x509.Certificate) []byte {
	return cert.Raw
}

func EncodeCertificates(certs []*x509.Certificate) []byte {
	var data []byte

	for _, cert := range certs {
		data2 := EncodeCertificate(cert)
		data = append(data, data2...)
	}

	return data
}

func EncodeCertificatesDER(certs []*x509.Certificate) []byte {
	var data []byte

	for _, cert := range certs {
		data2 := EncodeCertificateDER(cert)
		data = append(data, data2...)
	}

	return data
}
