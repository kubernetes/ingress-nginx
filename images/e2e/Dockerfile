# Copyright 2018 The Kubernetes Authors. All rights reserved.
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

FROM quay.io/kubernetes-ingress-controller/nginx-amd64:7b6e2dd312f1808e43fb39992ea814035557c7f3

ARG GOLANG_VERSION
ARG GOLANG_SHA

ARG RESTY_CLI_VERSION
ARG RESTY_CLI_SHA

ARG K8S_RELEASE
ARG ETCD_VERSION
ARG CHART_TESTING_VERSION

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

RUN apk add --no-cache \
  bash \
  ca-certificates \
  wget \
  make \
  gcc \
  git \
  musl-dev \
  perl \
  python \
  py-crcmod \
  py-pip \
  openssl

RUN set -eux; \
  apk add --no-cache --virtual .build-deps \
    g++ \
    pkgconfig \
    openssl \
    unzip \
    go \
  ; \
	export \
# set GOROOT_BOOTSTRAP such that we can actually build Go
		GOROOT_BOOTSTRAP="$(go env GOROOT)" \
# ... and set "cross-building" related vars to the installed system's values so that we create a build targeting the proper arch
# (for example, if our build host is GOARCH=amd64, but our build env/image is GOARCH=386, our build needs GOARCH=386)
		GOOS="$(go env GOOS)" \
		GOARCH="$(go env GOARCH)" \
		GOHOSTOS="$(go env GOHOSTOS)" \
		GOHOSTARCH="$(go env GOHOSTARCH)" \
	; \
# also explicitly set GO386 and GOARM if appropriate
# https://github.com/docker-library/golang/issues/184
	apkArch="$(apk --print-arch)"; \
	case "$apkArch" in \
		armhf) export GOARM='6' ;; \
		armv7) export GOARM='7' ;; \
		x86) export GO386='387' ;; \
	esac; \
	\
	wget -O go.tgz "https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz"; \
	echo "$GOLANG_SHA *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	cd /usr/local/go/src; \
	./make.bash; \
	\
	rm -rf \
# https://github.com/golang/go/blob/0b30cf534a03618162d3015c8705dd2231e34703/src/cmd/dist/buildtool.go#L121-L125
		/usr/local/go/pkg/bootstrap \
# https://golang.org/cl/82095
# https://github.com/golang/build/blob/e3fe1605c30f6a3fd136b561569933312ede8782/cmd/release/releaselet.go#L56
		/usr/local/go/pkg/obj \
	; \
	\
	export PATH="/usr/local/go/bin:$PATH"; \
	go version \
  ; \
  url="https://github.com/openresty/resty-cli/archive/v${RESTY_CLI_VERSION}.tar.gz"; \
  wget -O resty_cli.tgz "$url"; \
  echo "${RESTY_CLI_SHA} *resty_cli.tgz" | sha256sum -c -; \
  tar -C /tmp -xzf resty_cli.tgz; \
  rm resty_cli.tgz; \
  mv /tmp/resty-cli-${RESTY_CLI_VERSION}/bin/* /usr/local/bin/; \
  resty -V \
  ; \
  luarocks install luacheck; \
  luarocks install busted \
  ; \
  go get github.com/onsi/ginkgo/ginkgo; \
  go get golang.org/x/lint/golint \
  ; \
  apk del .build-deps;

RUN wget https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl

RUN wget https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/amd64/kube-apiserver -O /usr/local/bin/kube-apiserver \
  && chmod +x /usr/local/bin/kube-apiserver

RUN wget https://storage.googleapis.com/etcd/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz -O /tmp/etcd-${ETCD_VERSION}-linux-amd64.tar.gz \
  && mkdir -p /tmp/etcd-download \
  && tar xzvf /tmp/etcd-${ETCD_VERSION}-linux-amd64.tar.gz -C /tmp/etcd-download --strip-components=1 \
  && cp /tmp/etcd-download/etcd /usr/local/bin \
  && rm -rf /tmp/etcd-download

# Install a YAML Linter
ARG YAML_LINT_VERSION
RUN pip install "yamllint==$YAML_LINT_VERSION"

# Install Yamale YAML schema validator
ARG YAMALE_VERSION
RUN pip install "yamale==$YAMALE_VERSION"

RUN wget https://github.com/helm/chart-testing/releases/download/v${CHART_TESTING_VERSION}/chart-testing_${CHART_TESTING_VERSION}_linux_amd64.tar.gz \
    -O /tmp/ct-${CHART_TESTING_VERSION}-linux-amd64.tar.gz \
  && mkdir -p /tmp/ct-download \
  && tar xzvf /tmp/ct-${CHART_TESTING_VERSION}-linux-amd64.tar.gz -C /tmp/ct-download \
  && cp /tmp/ct-download/ct /usr/local/bin \
  && mkdir -p /etc/ct \
  && cp -R /tmp/ct-download/etc/* /etc/ct \
  && rm -rf /tmp/ct-download

RUN curl -sSL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

WORKDIR $GOPATH
