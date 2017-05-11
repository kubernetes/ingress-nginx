package ipport

import (
	"testing"
	
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	notCorsAnnotation = "ingress.kubernetes.io/enable-not-cors"
)

func TestParse(t *testing.T) {
	ap := NewParser()
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}
	
	testCases := []struct {
		annotations map[string]string
		expected    *IpPort
	}{
		{map[string]string{enableIpPortRequest: "true", ipportPort: "30085"}, &IpPort{true, "30085"}},
		{map[string]string{enableIpPortRequest: "false"}, &IpPort{}},
		{map[string]string{}, &IpPort{}},
		{nil, &IpPort{}},
	}
	
	ing := &extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.IngressSpec{},
	}
	
	for _, testCase := range testCases {
		ing.SetAnnotations(testCase.annotations)
		result, err := ap.Parse(ing)
		if err != nil {
			t.Errorf("parse result err: %v\n", err)
			return
		}
		
		res := result.(*IpPort)
		if res.Enable != testCase.expected.Enable || res.Port != testCase.expected.Port {
			t.Errorf("expected %t but returned %t, annotations: %s", testCase.expected, res, testCase.annotations)
		}
	}
}
