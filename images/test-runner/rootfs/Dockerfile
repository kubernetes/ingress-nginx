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
ARG BASE_IMAGE
ARG GOLANG_VERSION
ARG ETCD_VERSION

FROM golang:${GOLANG_VERSION}-alpine3.23 AS go
FROM registry.k8s.io/etcd:${ETCD_VERSION} AS etcd

FROM ${BASE_IMAGE}

ARG RESTY_CLI_VERSION
ARG RESTY_CLI_SHA
ARG LUAROCKS_VERSION
ARG LUAROCKS_SHA
ARG TARGETARCH
ARG K8S_RELEASE
ARG CHART_TESTING_VERSION
ARG HELM_VERSION
ARG YAMALE_VERSION
ARG YAML_LINT_VERSION
ARG GINKGO_VERSION
ARG GOLINT_VERSION


RUN set -eux; \
	if [ -e /etc/nsswitch.conf ]; then \
		grep '^hosts: files dns' /etc/nsswitch.conf; \
	else \
		echo 'hosts: files dns' > /etc/nsswitch.conf; \
	fi

COPY --from=go   /usr/local/go /usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

COPY --from=etcd /usr/local/bin/etcd /usr/local/bin/etcd

RUN echo "@testing https://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

RUN apk update && apk upgrade && apk add --no-cache \
  bash \
  ca-certificates \
  wget \
  make \
  gcc \
  git \
  musl-dev \
  perl \
  python3 \
  py-crcmod \
  py-pip \
  unzip \
  openssl \
  cfssl@testing \
  tzdata \
  libc6-compat \
  sqlite-dev

RUN go install -v github.com/onsi/ginkgo/v2/ginkgo@v${GINKGO_VERSION} \
    && go install golang.org/x/lint/golint@${GOLINT_VERSION}

RUN wget -qO /tmp/resty_cli.tgz \
  https://github.com/openresty/resty-cli/archive/v${RESTY_CLI_VERSION}.tar.gz \
  && echo "${RESTY_CLI_SHA} */tmp/resty_cli.tgz" | sha256sum -c - \
  && tar -C /tmp -xzf /tmp/resty_cli.tgz \
  && mv /tmp/resty-cli-${RESTY_CLI_VERSION}/bin/* /usr/local/bin/ \
  && resty -V \
  && rm -rf /tmp/*

RUN wget -qO /tmp/luarocks.tgz \
  https://github.com/luarocks/luarocks/archive/${LUAROCKS_VERSION}.tar.gz \
  && echo "${LUAROCKS_SHA} */tmp/luarocks.tgz" | sha256sum -c - \
  && tar -C /tmp -xzf /tmp/luarocks.tgz \
  && cd /tmp/luarocks* \
  && ./configure \
  && make install

RUN  luarocks install busted \
  && luarocks install luacheck

RUN wget -qO /usr/local/bin/kubectl \
  https://dl.k8s.io/release/${K8S_RELEASE}/bin/linux/${TARGETARCH}/kubectl \
  && chmod +x /usr/local/bin/kubectl

RUN wget -qO /usr/local/bin/kube-apiserver \
  https://dl.k8s.io/release/${K8S_RELEASE}/bin/linux/${TARGETARCH}/kube-apiserver \
  && chmod +x /usr/local/bin/kube-apiserver

RUN wget -qO /tmp/ct-${CHART_TESTING_VERSION}-linux-${TARGETARCH}.tar.gz \
  https://github.com/helm/chart-testing/releases/download/v${CHART_TESTING_VERSION}/chart-testing_${CHART_TESTING_VERSION}_linux_${TARGETARCH}.tar.gz \
  && mkdir -p /tmp/ct-download \
  && tar xzvf /tmp/ct-${CHART_TESTING_VERSION}-linux-${TARGETARCH}.tar.gz -C /tmp/ct-download \
  && rm /tmp/ct-${CHART_TESTING_VERSION}-linux-${TARGETARCH}.tar.gz \
  && cp /tmp/ct-download/ct /usr/local/bin \
  && mkdir -p /etc/ct \
  && cp -R /tmp/ct-download/etc/* /etc/ct \
  && rm -rf /tmp/*

RUN wget -qO /usr/local/bin/lj-releng \
  https://raw.githubusercontent.com/openresty/openresty-devel-utils/master/lj-releng \
  && chmod +x /usr/local/bin/lj-releng


RUN wget -qO /tmp/helm.tgz \
  https://get.helm.sh/helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz \
  && tar -C /tmp -xzf /tmp/helm.tgz \
  && cp /tmp/linux*/helm /usr/local/bin \
  && rm -rf /tmp/*

# Install a YAML Linter
# Pip not working. Check PR https://github.com/kubernetes/ingress-nginx/pull/10874
# RUN pip install --user "yamllint==$YAML_LINT_VERSION"
RUN apk update -U \
    && apk add yamllint \
    && yamllint --version

# Install Yamale YAML schema validator
# Commenting pip install yamale because broken cloudbuild https://github.com/kubernetes/ingress-nginx/pull/10885
# RUN pip install --user "yamale==$YAMALE_VERSION"
RUN wget https://github.com/23andMe/Yamale/archive/master.zip \
    && unzip -d /tmp master.zip \
    && cd /tmp/Yamale-master \
    && python setup.py install \
    && yamale -V

LABEL org.opencontainers.image.source=https://github.com/kubernetes/ingress-nginx

WORKDIR $GOPATH
