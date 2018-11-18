/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package defaultbackend

import (
	"crypto/tls"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const defaultBackend = "default backend - 404"

var _ = framework.IngressNginxDescribe("Default backend", func() {
	f := framework.NewDefaultFramework("default-backend")

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	It("should return 404 sending requests when only a default backend is running", func() {
		testCases := []struct {
			Name   string
			Host   string
			Scheme framework.RequestScheme
			Method string
			Path   string
			Status int
		}{
			{"basic HTTP GET request without host to path / should return 404", "", framework.HTTP, "GET", "/", http.StatusNotFound},
			{"basic HTTP GET request without host to path /demo should return 404", "", framework.HTTP, "GET", "/demo", http.StatusNotFound},
			{"basic HTTPS GET request without host to path / should return 404", "", framework.HTTPS, "GET", "/", http.StatusNotFound},
			{"basic HTTPS GET request without host to path /demo should return 404", "", framework.HTTPS, "GET", "/demo", http.StatusNotFound},

			{"basic HTTP POST request without host to path / should return 404", "", framework.HTTP, "POST", "/", http.StatusNotFound},
			{"basic HTTP POST request without host to path /demo should return 404", "", framework.HTTP, "POST", "/demo", http.StatusNotFound},
			{"basic HTTPS POST request without host to path / should return 404", "", framework.HTTPS, "POST", "/", http.StatusNotFound},
			{"basic HTTPS POST request without host to path /demo should return 404", "", framework.HTTPS, "POST", "/demo", http.StatusNotFound},

			{"basic HTTP GET request to host foo.bar.com and path / should return 404", " foo.bar.com", framework.HTTP, "GET", "/", http.StatusNotFound},
			{"basic HTTP GET request to host foo.bar.com and path /demo should return 404", " foo.bar.com", framework.HTTP, "GET", "/demo", http.StatusNotFound},
			{"basic HTTPS GET request to host foo.bar.com and path / should return 404", " foo.bar.com", framework.HTTPS, "GET", "/", http.StatusNotFound},
			{"basic HTTPS GET request to host foo.bar.com and path /demo should return 404", " foo.bar.com", framework.HTTPS, "GET", "/demo", http.StatusNotFound},

			{"basic HTTP POST request to host foo.bar.com and path / should return 404", " foo.bar.com", framework.HTTP, "POST", "/", http.StatusNotFound},
			{"basic HTTP POST request to host foo.bar.com and path /demo should return 404", " foo.bar.com", framework.HTTP, "POST", "/demo", http.StatusNotFound},
			{"basic HTTPS POST request to host foo.bar.com and path / should return 404", " foo.bar.com", framework.HTTPS, "POST", "/", http.StatusNotFound},
			{"basic HTTPS POST request to host foo.bar.com and path /demo should return 404", " foo.bar.com", framework.HTTPS, "POST", "/demo", http.StatusNotFound},
		}

		for _, test := range testCases {
			By(test.Name)

			request := gorequest.New()
			var cm *gorequest.SuperAgent

			switch test.Scheme {
			case framework.HTTP:
				cm = request.CustomMethod(test.Method, f.IngressController.HTTPURL)
			case framework.HTTPS:
				cm = request.CustomMethod(test.Method, f.IngressController.HTTPSURL)
				// the default backend uses a self generated certificate
				cm.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
			default:
				Fail("Unexpected request scheme")
			}

			if test.Host != "" {
				cm.Set("Host", test.Host)
			}

			resp, _, errs := cm.End()
			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(test.Status))
		}
	})
})
