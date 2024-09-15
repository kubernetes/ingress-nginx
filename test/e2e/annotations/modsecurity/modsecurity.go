/*
Copyright 2018 The Kubernetes Authors.

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

package modsecurity

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	modSecurityFooHost = "modsecurity.foo.com"
	defaultSnippet     = `SecRuleEngine On
		SecRequestBodyAccess On
		SecAuditEngine RelevantOnly
		SecAuditLogParts ABIJDEFHZ
		SecAuditLog /dev/stdout
		SecAuditLogType Serial
		SecRule REQUEST_HEADERS:User-Agent \"block-ua\" \"log,deny,id:107,status:403,msg:\'UA blocked\'\"`
)

var _ = framework.DescribeAnnotation("modsecurity owasp", func() {
	f := framework.NewDefaultFramework("modsecuritylocation")
	if framework.IsCrossplane() {
		return
	}

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should enable modsecurity", func() {
		host := modSecurityFooHost
		nameSpace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "modsecurity on;") &&
					strings.Contains(server, "modsecurity_rules_file /etc/nginx/modsecurity/modsecurity.conf;")
			})
	})

	ginkgo.It("should enable modsecurity with transaction ID and OWASP rules", func() {
		host := modSecurityFooHost
		nameSpace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity":         "true",
			"nginx.ingress.kubernetes.io/enable-owasp-core-rules":    "true",
			"nginx.ingress.kubernetes.io/modsecurity-transaction-id": "modsecurity-$request_id",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "modsecurity on;") &&
					strings.Contains(server, "modsecurity_rules_file /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf;") &&
					strings.Contains(server, "modsecurity_transaction_id \"modsecurity-$request_id\";")
			})
	})

	ginkgo.It("should disable modsecurity", func() {
		host := modSecurityFooHost
		nameSpace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity": "false",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "modsecurity on;")
			})
	})

	ginkgo.It("should enable modsecurity with snippet", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		host := modSecurityFooHost
		nameSpace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity":  "true",
			"nginx.ingress.kubernetes.io/modsecurity-snippet": "SecRuleEngine On",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "modsecurity on;") &&
					strings.Contains(server, "SecRuleEngine On")
			})
	})

	ginkgo.It("should enable modsecurity without using 'modsecurity on;'", func() {
		f.SetNginxConfigMapData(map[string]string{
			"enable-modsecurity": "true",
		},
		)

		host := modSecurityFooHost
		nameSpace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "modsecurity on;") &&
					!strings.Contains(server, "modsecurity_rules_file /etc/nginx/modsecurity/modsecurity.conf;")
			})
	})

	ginkgo.It("should disable modsecurity using 'modsecurity off;'", func() {
		f.SetNginxConfigMapData(map[string]string{
			"enable-modsecurity": "true",
		},
		)

		host := modSecurityFooHost
		nameSpace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity": "false",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "modsecurity off;")
			})
	})

	ginkgo.It("should enable modsecurity with snippet and block requests", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		host := modSecurityFooHost
		nameSpace := f.Namespace

		snippet := defaultSnippet

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity":  "true",
			"nginx.ingress.kubernetes.io/modsecurity-snippet": snippet,
		}
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "load_module, lua_package, _by_lua, location, root, {, }")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "modsecurity on;") &&
					strings.Contains(server, "SecRuleEngine On")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "block-ua").
			Expect().
			Status(http.StatusForbidden)
	})

	ginkgo.It("should enable modsecurity globally and with modsecurity-snippet block requests", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		host := modSecurityFooHost
		nameSpace := f.Namespace

		snippet := defaultSnippet

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/modsecurity-snippet": snippet,
		}
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "load_module, lua_package, _by_lua, location, root, {, }")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.UpdateNginxConfigMapData("enable-modsecurity", "true")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "SecRuleEngine On")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "block-ua").
			Expect().
			Status(http.StatusForbidden)
	})

	ginkgo.It("should enable modsecurity when enable-owasp-modsecurity-crs is set to true", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		f.UpdateNginxConfigMapData("enable-modsecurity", "true")
		f.UpdateNginxConfigMapData("enable-owasp-modsecurity-crs", "true")

		host := modSecurityFooHost
		nameSpace := f.Namespace

		snippet := defaultSnippet

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/modsecurity-snippet": snippet,
		}
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "load_module, lua_package, _by_lua, location, root, {, }")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "SecRuleEngine On")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "block-ua").
			Expect().
			Status(http.StatusForbidden)
	})

	ginkgo.It("should enable modsecurity through the config map", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()
		host := modSecurityFooHost
		nameSpace := f.Namespace

		snippet := `SecRequestBodyAccess On
		SecAuditEngine RelevantOnly
		SecAuditLogParts ABIJDEFHZ
		SecAuditLog /dev/stdout
		SecAuditLogType Serial
		SecRule REQUEST_HEADERS:User-Agent \"block-ua\" \"log,deny,id:107,status:403,msg:\'UA blocked\'\"`

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/modsecurity-snippet": snippet,
		}
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "load_module, lua_package, _by_lua, location, root, {, }")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		expectedComment := "SecRuleEngine On"
		f.UpdateNginxConfigMapData("enable-modsecurity", "true")
		f.UpdateNginxConfigMapData("enable-owasp-modsecurity-crs", "true")
		f.UpdateNginxConfigMapData("modsecurity-snippet", expectedComment)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "SecRequestBodyAccess On")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "block-ua").
			Expect().
			Status(http.StatusForbidden)
	})

	ginkgo.It("should enable modsecurity through the config map but ignore snippet as disabled by admin", func() {
		host := modSecurityFooHost
		nameSpace := f.Namespace

		f.UpdateNginxConfigMapData("annotations-risk-level", "Critical") // To enable snippet configurations
		defer f.UpdateNginxConfigMapData("annotations-risk-level", "High")

		snippet := `SecRequestBodyAccess On
		SecAuditEngine RelevantOnly
		SecAuditLogParts ABIJDEFHZ
		SecAuditLog /dev/stdout
		SecAuditLogType Serial
		SecRule REQUEST_HEADERS:User-Agent \"block-ua\" \"log,deny,id:107,status:403,msg:\'UA blocked\'\"`

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/modsecurity-snippet": snippet,
		}
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "load_module, lua_package, _by_lua, location, root, {, }")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		expectedComment := "SecRuleEngine On"

		f.SetNginxConfigMapData(map[string]string{
			"enable-modsecurity":           "true",
			"enable-owasp-modsecurity-crs": "true",
			"allow-snippet-annotations":    "false",
			"modsecurity-snippet":          expectedComment,
		})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "block-ua")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "block-ua").
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should disable default modsecurity conf setting when modsecurity-snippet is specified", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		host := modSecurityFooHost
		nameSpace := f.Namespace

		snippet := `SecRuleEngine On
		SecRequestBodyAccess On
		SecAuditEngine RelevantOnly
		SecAuditLogParts ABIJDEFHZ
		SecAuditLogType Concurrent
		SecAuditLog /var/tmp/modsec_audit.log
		SecAuditLogStorageDir /var/tmp/
		SecRule REQUEST_HEADERS:User-Agent \"block-ua\" \"log,deny,id:107,status:403,msg:\'UA blocked\'\"`

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity":  "true",
			"nginx.ingress.kubernetes.io/modsecurity-snippet": snippet,
		}
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "load_module, lua_package, _by_lua, location, root, {, }")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "modsecurity_rules_file /etc/nginx/modsecurity/modsecurity.conf;") &&
					strings.Contains(server, "SecAuditLog /var/tmp/modsec_audit.log")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "block-ua").
			Expect().
			Status(http.StatusForbidden)
	})
})
