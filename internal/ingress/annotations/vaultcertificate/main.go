/*

We create this annotation in order for us to be able to exploit the use of vault stored certificates

*/

package backendCertVaultPath

import (
	"errors"
	"regexp"
	"strings"

	networking "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

var VaultCertificate string

const EmptyVaultPath = ""

var (
	validVaultUrl = regexp.MustCompile(`^/(?:[\w-]+)[\w\*\.\/\-\_]+$`)
)

type backendCertVaultPath struct {
	r resolver.Resolver
}

// NewParser creates a new backend protocol annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return backendCertVaultPath{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to indicate the backend protocol.
func (a backendCertVaultPath) Parse(ing *networking.Ingress) (interface{}, error) {
	if ing.GetAnnotations() == nil {
		return EmptyVaultPath, nil
	}

	VaultCertificate, err := parser.GetStringAnnotation("tls-cert-vault", ing)
	if err != nil {
		return EmptyVaultPath, nil
	}

	VaultCertificate = strings.TrimSpace(VaultCertificate)
	if !validVaultUrl.MatchString(VaultCertificate) {
		klog.Errorf("URL %v is not a valid value for the tls-cert-vault annotation. Regex rule is: %v", VaultCertificate, validVaultUrl)
		err := errors.New("not a valid value for the tls-cert-vault annotation")
		return EmptyVaultPath, err
	}

	return VaultCertificate, nil
}
