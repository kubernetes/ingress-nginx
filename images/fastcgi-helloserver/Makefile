# Copyright 2020 The Kubernetes Authors.
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

# Docker image for e2e testing.

# Use the 0.0 tag for testing, it shouldn't clobber any release builds
TAG ?= 0.0

REGISTRY ?= ingress-controller
DOCKER ?= docker

IMGNAME = fastcgi-helloserver
IMAGE = $(REGISTRY)/$(IMGNAME)

PKG=k8s.io/ingress-nginx/images/fastcgi-helloserver

container: clean build
	$(DOCKER) buildx build \
		--load \
		--platform linux/amd64 \
		-t $(IMAGE):$(TAG) rootfs

build: clean
	CGO_ENABLED=0 go build -a -installsuffix cgo \
		-ldflags "-s -w" \
		-o rootfs/fastcgi-helloserver ${PKG}/...

clean:
	$(DOCKER) rmi -f $(IMAGE):$(TAG) || true
