# Copyright 2022 The Kubernetes Authors. All rights reserved.
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

.DEFAULT_GOAL:=melange

# set default shell
SHELL=/bin/bash -o pipefail -o errexit
DIR:=$(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))

# 0.0.0 shouldn't clobber any released builds
TAG ?= 0.0
REGISTRY ?= gcr.io/k8s-staging-ingress-nginx

IMAGE ?= $(REGISTRY)/controller

ARCH := $(shell uname -m)
MELANGE_DIR ?= melange
APKO_DIR ?= apko
MELANGE ?= docker run --rm --privileged -w /work -v "${PWD}":/work distroless.dev/melange:latest
MELANGE_DETACHED ?= docker run -d -w /work --rm --privileged -v "${PWD}":/work distroless.dev/melange:latest
APKO ?= docker run --rm -w /work -v "${PWD}":/work ko.local:ca10c03b79422ee648991e16865b62fd26bac97ed4238e7890d5097896af1840
KEY ?= melange.rsa
REPO ?= packages
TEMPLATE ?= melange/nginx-templates.json
MELANGE_OPTS ?= -k ${KEY}.pub --signing-key ${KEY} --arch ${ARCHS}
MELANGE_INGRESS_OPT ?= -k ${KEY}.pub --signing-key ${KEY} --arch ${ARCHS} --empty-workspace
APKO_OPTS ?= -k ${KEY}.pub --debug --sbom=false --build-arch ${ARCHS} ${APKO_DIR}/${FILE}.yaml
KEY ?= melange.rsa
REPO ?= $(shell pwd)/packages
ARCHS ?="amd64,arm64,arm/v6,arm/v7,s390x"

define build-package
	docker run $(2) --rm --privileged -v "${PWD}":/work distroless.dev/melange:latest build ${MELANGE_DIR}/$(1).yaml ${MELANGE_OPTS} --empty-workspace --template '$(shell cat ${TEMPLATE})'
endef

keygen: ## Generate Key pair for use with signing apks
	docker run --rm --privileged -v "${PWD}":/work distroless.dev/melange:latest keygen

.PHONY:	melange
melange: ## Build melange $FILE
	${MELANGE} build ${MELANGE_DIR}/${FILE}.yaml ${MELANGE_OPTS} --template '$(shell cat ${TEMPLATE})'

ingress-build:
	${MELANGE} build ${MELANGE_DIR}/${FILE}.yaml ${MELANGE_INGRESS_OPT} --template '$(shell cat ${TEMPLATE})'

nginx-build:
	${MELANGE} build ${MELANGE_DIR}/${FILE}.yaml --source-dir ${MELANGE_DIR}/${FILE}  ${MELANGE_OPTS} --template '$(shell cat ${TEMPLATE})'

apko-build: ## Build an apko pipeline with $KEY and $FILE
	${APKO} build ${APKO_OPTS} $(IMAGE):$(TAG) $(IMAGE)-$(TAG).tar

apko-push: ## Push apko built container $IMAGE:$TAG to $REGISTRY
	${APKO} publish ${APKO_OPTS} $(IMAGE):$(TAG)

load: ## Load apko built image into docker
	docker load < $(IMAGE)-$(TAG).tar
	docker tag $(IMAGE):$(TAG) $(REGISTRY)/$(IMAGE):$(TAG)
	docker push $(REGISTRY)/$(IMAGE):$(TAG)
	rm $(IMAGE)-$(TAG).tar

build-all: clean-packages all-packages nginx-package ingress-packages ## Fresh build of all melange pipelines and apko files, default is all $ARCHS

nginx-test: ## Start $IMAGE:$TAG container and drop into bash shell
	docker run --rm -it --entrypoint bash --env-file .env $(REGISTRY)/$(IMAGE):$(TAG)

shell: ## Start Alpine base container, mount PWD and drop into sh
	docker run -it --rm -v "${PWD}/..":/work --env-file .env --group-add www-data --entrypoint sh distroless.dev/alpine-base:latest

check_clean:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

clean-packages: check_clean ## Deletes /packages directory, this will take 4 hours to rebuild, are you sure?
	echo "deleting everything"
	rm -rf packages/ && mkdir -p packages/

# order matters here, jaeger needs yaml etc
all-packages:
	$(call  build-package,mimaloc,"-d")
	$(call  build-package,ssdeep,"-d")
	$(call  build-package,opentracing,)
	$(call  build-package,msgpack,)
	$(call  build-package,datadog,)
	$(call  build-package,yaml,)
	$(call  build-package,zipkin,)
	$(call  build-package,jaeger,)
	$(call  build-package,luajit,)
	$(call  build-package,lua-resty-core,)
	$(call  build-package,modsecurity,)
	$(call  build-package,lua-resty-upload,"-d")
	$(call  build-package,lua-resty-string,"-d")
	$(call  build-package,lua-resty-redis,"-d")
	$(call  build-package,lua-resty-memcached,"-d")
	$(call  build-package,lua-resty-lock,"-d")
	$(call  build-package,lua-resty-ipmatcher,"-d")
	$(call  build-package,lua-resty-http,"-d")
	$(call  build-package,lua-resty-global-throttle,"-d")
	$(call  build-package,lua-resty-dns,"-d")
	$(call  build-package,lua-resty-cookie,"-d")
	$(call  build-package,lua-resty-cache,"-d")
	$(call  build-package,lua-resty-balancer,"-d")
	$(call  build-package,lua-cjson,"-d")

nginx-package:
	$(call  build-package,nginx)

ingress-packages:
	$(call  build-package,ingress)
	$(call  build-package,dbg)
	$(call  build-package,wait-shutdown)
