# Copyright 2025 The Kubernetes Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

NAME ?=

REGISTRY ?= us-central1-docker.pkg.dev/k8s-staging-images/ingress-nginx
REPOSITORY ?= $(REGISTRY)/charts

.PHONY: helm
helm:
	command -v helm || go install helm.sh/helm/v4/cmd/helm@v4.0.4

.PHONY: package
package: helm
	helm package $(NAME)

.PHONY: push
push: package
	helm push $(NAME)-*.tgz oci://$(REPOSITORY)
