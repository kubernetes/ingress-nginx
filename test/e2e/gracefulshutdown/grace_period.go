/*
Copyright 2020 The Kubernetes Authors.

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

package gracefulshutdown

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Shutdown] Grace period shutdown", func() {
	f := framework.NewDefaultFramework("shutdown-grace-period")

	ginkgo.It("/healthz should return status code 500 during shutdown grace period", func() {

		f.NewSlowEchoDeployment()

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := []string{}
			for _, v := range deployment.Spec.Template.Spec.Containers[0].Args {
				if strings.Contains(v, "--shutdown-grace-period") {
					continue
				}

				args = append(args, v)
			}

			args = append(args, "--shutdown-grace-period=90")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			cmds := []string{"/wait-shutdown"}
			deployment.Spec.Template.Spec.Containers[0].Lifecycle.PreStop.Exec.Command = cmds
			grace := int64(3600)
			deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &grace
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})

		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

		ip := f.GetNginxPodIP()

		err = f.VerifyHealthz(ip, http.StatusOK)
		assert.Nil(ginkgo.GinkgoT(), err)

		result := make(chan []error)
		go func(c chan []error) {
			defer ginkgo.GinkgoRecover()
			errors := []error{}

			framework.Sleep(60 * time.Second)

			err = f.VerifyHealthz(ip, http.StatusInternalServerError)
			if err != nil {
				errors = append(errors, err)
			}

			c <- errors
		}(result)

		f.ScaleDeploymentToZero("nginx-ingress-controller")

		for _, err := range <-result {
			assert.Nil(ginkgo.GinkgoT(), err)
		}

	})
})
