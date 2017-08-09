package controller

import (
	"testing"

	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/ingress/core/pkg/ingress"
)

func buildGenericControllerForCertTest() *GenericController {
	return &GenericController{
		cfg: &Configuration{},
	}
}

func buildTestIngressRuleForCertTest() extensions.Ingress {
	return extensions.Ingress{
		Spec: extensions.IngressSpec{
			Rules: []extensions.IngressRule{
				extensions.IngressRule{
					Host: "hostWithTLS",
				},
				extensions.IngressRule{
					Host: "hostWithoutTLS",
				},
			},
			TLS: []extensions.IngressTLS{
				extensions.IngressTLS{
					Hosts: []string{"hostWithTLS"},
				},
			},
		},
	}
}

func buildTestServersForCertTest() map[string]*ingress.Server {
	return map[string]*ingress.Server{
		"hostWithTLS": &ingress.Server{
			Hostname: "hostWithTLS",
		},
		"hostWithoutTLS": &ingress.Server{
			Hostname: "hostWithoutTLS",
		},
	}
}

func TestConfigureTLSforIng(t *testing.T) {
	ic := buildGenericControllerForCertTest()

	testIng := buildTestIngressRuleForCertTest()
	testServers := buildTestServersForCertTest()

	defaultPemFileName := "defaultPemFileName"
	defaultPemSHA := "defaultPemSHA"

	ic.configureTLSforIng(&testIng, testServers, defaultPemFileName, defaultPemSHA)
	if testServers["hostWithTLS"].SSLCertificate != defaultPemFileName {
		t.Errorf("SSLCertificate set to %s instead of %s", testServers["hostWithTLS"].SSLCertificate, defaultPemFileName)
	}
	if testServers["hostWithoutTLS"].SSLCertificate != "" {
		t.Errorf("SSLCertificate set to %s instead of being empty", testServers["hostWithoutTLS"].SSLCertificate)
	}

	ic.cfg.AlwaysEnableTLS = true
	ic.configureTLSforIng(&testIng, testServers, defaultPemFileName, defaultPemSHA)
	if testServers["hostWithTLS"].SSLCertificate != defaultPemFileName {
		t.Errorf("SSLCertificate set to %s instead of %s", testServers["hostWithTLS"].SSLCertificate, defaultPemFileName)
	}
	if testServers["hostWithoutTLS"].SSLCertificate != defaultPemFileName {
		t.Errorf("SSLCertificate set to %s instead of %s", testServers["hostWithoutTLS"].SSLCertificate, defaultPemFileName)
	}
}
