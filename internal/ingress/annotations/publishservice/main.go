package publishservice

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	networking "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

var (
	publishServiceRegex = regexp.MustCompile(`basic|digest`)
)

type publishservice struct {
	r resolver.Resolver
}

// NewParser creates a new publish-service annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return publishservice{r}
}

// Parse parses the annotations contained in the Ingress to use
// a custom Service for obtaining published addresses
func (ps publishservice) Parse(ing *networking.Ingress) (interface{}, error) {
	key, err := parser.GetStringAnnotation("publish-service", ing)
	if err != nil {
		return nil, err
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "error reading Service name from annotation publish-service"),
		}
	}

	if ns == "" {
		ns = ing.Namespace
	}

	key = fmt.Sprintf("%v/%v", ns, name)
	_, err = ps.r.GetService(key)
	if err != nil {
		return nil, ing_errors.LocationDenied{
			Reason: errors.Wrapf(err, "unexpected error reading Service %v", key),
		}
	}

	return key, nil
}
