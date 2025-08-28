/*
Copyright 2025 The Kubernetes Authors.

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

package settings

import (
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("ssl-session-tickets", func() {
	f := framework.NewDefaultFramework("ssl-session-tickets")

	ginkgo.It("should have default ssl_session_tickets value", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "ssl_session_tickets off;")
		})
	})

	ginkgo.It("should set ssl_session_tickets value", func() {
		f.UpdateNginxConfigMapData("ssl-session-tickets", "true")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "ssl_session_tickets on;")
		})
	})

	ginkgo.It("should set ssl_session_tickets and ssl_session_ticket_key values", func() {
		f.UpdateNginxConfigMapData("ssl-session-tickets", "true")
		f.UpdateNginxConfigMapData("ssl-session-ticket-key", "WW9gcPHgfcrw6DNqY5VE2NjM6gtgUhJ4Vn6ZwRGi/7+A9TNFa4Fvfe1cmlPec9bxDoenN70aMBeZBlcrKshnKT4WJxFNLCuTHhfn4loTOEo=")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "ssl_session_tickets on;") &&
				strings.Contains(cfg, "ssl_session_ticket_key /etc/ingress-controller/tickets.key;")
		})
	})
})
