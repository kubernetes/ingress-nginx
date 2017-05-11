package ipport

import (
	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	enableIpPortRequest = "ingress.kubernetes.io/enable-ipport"
	ipportPort          = "ingress.kubernetes.io/ipport"
)

type IpPort struct {
	Enable bool   `json:"enable"`
	Port   string `json:"port"`
}

type ipPort struct {

}

// NewParser creates a new CORS annotation parser
func NewParser() parser.IngressAnnotation {
	return ipPort{}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the location/s should allows CORS
func (a ipPort) Parse(ing *extensions.Ingress) (interface{}, error) {
	enable, err := parser.GetBoolAnnotation(enableIpPortRequest, ing)
	if err != nil {
		return nil, err
	}

	port, err := parser.GetStringAnnotation(ipportPort, ing)
	if err != nil {
		return nil, err
	}

	return &IpPort{
		Enable: enable,
		Port:   port,
	}, nil
}
