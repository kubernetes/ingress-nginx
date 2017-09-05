# Copyright 2017 The Kubernetes Authors. All rights reserved.
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

# Build the default backend binary or image for amd64, arm, arm64 and ppc64le
#
# Usage:
# 	[PREFIX=gcr.io/google_containers/dummy-ingress-controller] [ARCH=amd64] [TAG=1.1] make (server|container|push)

all: push

TAG=0.1
PREFIX?=bprashanth/dummy-ingress-controller
ARCH?=amd64
GOLANG_VERSION=1.6
TEMP_DIR:=$(shell mktemp -d)

server: server.go
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) GOARM=6 godep go build -a -installsuffix cgo -ldflags '-w' -o server ./server.go

container: server
	docker build --pull -t $(PREFIX)-$(ARCH):$(TAG) .

push: container
	gcloud docker -- push $(PREFIX)-$(ARCH):$(TAG)

clean:
	rm -f server

