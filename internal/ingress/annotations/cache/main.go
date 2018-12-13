package cache

import (
	"fmt"
	"os"
	"github.com/golang/glog"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type Config struct {
	Enabled         bool   `json:"enabled"`
	Name            string `json:"name"`
	AddStatusHeader bool   `json:"addStatusHeader"`
	InactiveTimeout string `json:"inactiveTimeout"`
	Levels          string `json:"levels"`
	MaxSize         string `json:"maxSize"`
	Methods         string `json:"methods"`
	MinUses         int    `json:"minUses"`
	Path            string `json:"path"`
	Revalidate      string `json:"revalidate"`
	TTL             string `json:"ttl"`
	UseStale        string `json:"useStale"`
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
	if c1.Name != c2.Name {
		return false
	}
	if c1.AddStatusHeader != c2.AddStatusHeader {
		return false
	}
	if c1.InactiveTimeout != c2.InactiveTimeout {
		return false
	}
	if c1.Levels != c2.Levels {
		return false
	}
	if c1.MaxSize != c2.MaxSize {
		return false
	}
	if c1.Methods != c2.Methods {
		return false
	}
	if c1.MinUses != c2.MinUses {
		return false
	}
	if c1.MaxSize != c2.MaxSize {
		return false
	}
	if c1.Path != c2.Path {
		return false
	}
	if c1.Revalidate != c2.Revalidate {
		return false
	}
	if c1.UseStale != c2.UseStale {
		return false
	}
	if c1.TTL != c2.TTL {
		return false
	}
 	return true
}

type cache struct {
	r resolver.Resolver
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return cache{r}
}

func (a cache) Parse(ing *extensions.Ingress) (interface{}, error) {
	e, err := parser.GetBoolAnnotation("enable-cache", ing)
	if err != nil {
		e = false
	}

	l, err := parser.GetStringAnnotation("cache-levels", ing)
	if err != nil {
		l = "1:2"
	}

	ms, err := parser.GetStringAnnotation("cache-max-size", ing)
	if err != nil {
		ms = "100m"
	}

	m, err := parser.GetStringAnnotation("cache-methods", ing)
	if err != nil {
		m = "GET HEAD"
	}

	it, err := parser.GetStringAnnotation("cache-inactive-timeout", ing)
	if err != nil {
		it = "60m"
	}

	us, err := parser.GetStringAnnotation("cache-use-stale", ing)
	if err != nil {
		us = "error timeout updating http_500 http_502 http_503 http_504"
	}

	r, err := parser.GetStringAnnotation("cache-revalidate", ing)
	if err != nil {
		r = "on"
	}

 	mu, err := parser.GetIntAnnotation("cache-min-uses", ing)
	if err != nil {
		mu = 3
	}

 	ash, err := parser.GetBoolAnnotation("cache-add-status-header", ing)
	if err != nil {
		ash = true
	}

 	ttl, err := parser.GetStringAnnotation("cache-valid-for", ing)
	if err != nil {
		ttl = "10m"
	}

	p := fmt.Sprintf("/var/lib/nginx/cache/%v/%v", ing.Namespace, ing.Name)
	n := fmt.Sprintf("%v-%v", ing.Namespace, ing.Name)

	err = os.MkdirAll(p, 0777)

	if err != nil {
		glog.Errorf("unexpected error creating cache directory %v: %v", p, err)
	}

	glog.Info("Created cache directory %v", p)

	return &Config{e, n, ash, it, l, ms, m, mu, p, r, ttl, us}, nil

}
