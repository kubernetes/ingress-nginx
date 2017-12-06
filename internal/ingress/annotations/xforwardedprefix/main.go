package xforwardedprefix

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type xforwardedprefix struct {
	r resolver.Resolver
}

// NewParser creates a new xforwardedprefix annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return xforwardedprefix{r}
}

// Parse parses the annotations contained in the ingress rule
// used to add an x-forwarded-prefix header to the request
func (cbbs xforwardedprefix) Parse(ing *extensions.Ingress) (interface{}, error) {
	return parser.GetBoolAnnotation("x-forwarded-prefix", ing)
}
