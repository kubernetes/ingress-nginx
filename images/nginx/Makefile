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

BUILDER ?= ingress-nginx
PLATFORMS ?= linux/amd64,linux/arm,linux/arm64
REGISTRY ?= us-central1-docker.pkg.dev/k8s-staging-images/ingress-nginx
IMAGE ?= $(REGISTRY)/nginx
TAG ?= $(shell cat TAG)

.PHONY: builder
builder:
	docker buildx create --name $(BUILDER) --bootstrap || :
	docker buildx inspect $(BUILDER)

.PHONY: build
build: builder
	docker buildx build \
		--builder $(BUILDER) \
		--platform $(PLATFORMS) \
		rootfs \
		--tag $(IMAGE):$(TAG)

# Pushing in the `build` target does not work as authentication times out after one hour.
#
# Therefore we need to build and push in separate commands.
.PHONY: push
push: build
	docker buildx build \
		--builder $(BUILDER) \
		--platform $(PLATFORMS) \
		rootfs \
		--tag $(IMAGE):$(TAG) \
		--push

.PHONY: clean
clean:
	docker buildx rm $(BUILDER) || :
