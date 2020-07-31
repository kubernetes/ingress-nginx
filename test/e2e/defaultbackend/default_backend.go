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
	"net/http"

	"github.com/gavv/httpexpect/v2"
	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Default Backend]", func() {
	f := framework.NewDefaultFramework("default-backend")

	ginkgo.It("should return 404 sending requests when only a default backend is running", func() {
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

		framework.Sleep()

		for _, test := range testCases {
			ginkgo.By(test.Name)

			var req *httpexpect.Request

			switch test.Scheme {
			case framework.HTTP:
				req = f.HTTPTestClient().Request(test.Method, test.Path)
				req.WithURL(f.GetURL(framework.HTTP) + test.Path)
			case framework.HTTPS:
				req = f.HTTPTestClient().Request(test.Method, test.Path)
				req.WithURL(f.GetURL(framework.HTTPS) + test.Path)
			default:
				ginkgo.Fail("Unexpected request scheme")
			}

			if test.Host != "" {
				req.WithHeader("Host", test.Host)
			}

			req.Expect().
				Status(test.Status)
		}
	})

	ginkgo.It("enables access logging for default backend", func() {
		// TODO: fix
		ginkgo.Skip("enable-access-log-for-default-backend")

		f.UpdateNginxConfigMapData("enable-access-log-for-default-backend", "true")

		f.HTTPTestClient().
			GET("/somethingOne").
			WithHeader("Host", "foo").
			Expect().
			Status(http.StatusNotFound)

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, "/somethingOne")
	})

	ginkgo.It("disables access logging for default backend", func() {
		// TODO: fix
		ginkgo.Skip("enable-access-log-for-default-backend")

		// enable-access-log-for-default-backend is false by default, setting the value to false do not trigger a reload
		f.UpdateNginxConfigMapData("enable-access-log-for-default-backend", "true")
		f.UpdateNginxConfigMapData("enable-access-log-for-default-backend", "false")

		f.HTTPTestClient().
			GET("/somethingTwo").
			WithHeader("Host", "bar").
			Expect().
			Status(http.StatusNotFound)

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.NotContains(ginkgo.GinkgoT(), logs, "/somethingTwo")
	})
})
