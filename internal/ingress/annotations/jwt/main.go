package jwt

import (
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type Config struct {
	Enabled         bool   `json:"enabled"`
	ResponseContentType string `json:"responseContentType"`
	ResponseData      string `json:"responseData"`
}

func (c1 *Config) Equal(c2 *Config) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
 	if c1.Enabled != c2.Enabled {
		return false
	}
	if c1.ResponseContentType != c2.ResponseContentType {
		return false
	}
 	return true
}

type jwt struct {
	r resolver.Resolver
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return jwt{r}
}

func (a jwt) Parse(ing *extensions.Ingress) (interface{}, error) {
	e, err := parser.GetBoolAnnotation("enable-jwt", ing)
	if err != nil {
		e = false
	}

	rc, err := parser.GetStringAnnotation("jwt-response-content-type", ing)
	if err != nil {
		rc = "application/json"
	}

	rd, err := parser.GetStringAnnotation("jwt-response-data", ing)
	if err != nil {
		rd = "{\\\"status\\\": \\\"unauthorized\\\"}"
	}

	return &Config{e,rc,rd}, nil

}
