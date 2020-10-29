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

FROM golang:${GOLANG_VERSION}-alpine as GO
FROM k8s.gcr.io/etcd:${ETCD_VERSION} as etcd

FROM ${BASE_IMAGE}

RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

COPY --from=GO   /usr/local/go /usr/local/go
COPY --from=etcd /usr/local/bin/etcd /usr/local/bin/etcd

RUN echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

RUN apk add --no-cache \
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
  cfssl@testing

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN go get github.com/onsi/ginkgo/ginkgo golang.org/x/lint/golint

ARG RESTY_CLI_VERSION
ARG RESTY_CLI_SHA

RUN wget -O /tmp/resty_cli.tgz \
  https://github.com/openresty/resty-cli/archive/v${RESTY_CLI_VERSION}.tar.gz \
  && echo "${RESTY_CLI_SHA} */tmp/resty_cli.tgz" | sha256sum -c - \
  && tar -C /tmp -xzf /tmp/resty_cli.tgz \
  && mv /tmp/resty-cli-${RESTY_CLI_VERSION}/bin/* /usr/local/bin/ \
  && resty -V \
  && rm -rf /tmp/*

ARG LUAROCKS_VERSION
ARG LUAROCKS_SHA

RUN wget -O /tmp/luarocks.tgz \
  https://github.com/luarocks/luarocks/archive/v${LUAROCKS_VERSION}.tar.gz \
  && echo "${LUAROCKS_SHA} */tmp/luarocks.tgz" | sha256sum -c - \
  && tar -C /tmp -xzf /tmp/luarocks.tgz \
  && cd /tmp/luarocks* \
  && ./configure \
  && make install

RUN  luarocks install busted \
  && luarocks install luacheck

ARG TARGETARCH

ARG K8S_RELEASE

RUN wget -O /usr/local/bin/kubectl \
  https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/${TARGETARCH}/kubectl \
  && chmod +x /usr/local/bin/kubectl

RUN wget -O /usr/local/bin/kube-apiserver \
  https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/${TARGETARCH}/kube-apiserver \
  && chmod +x /usr/local/bin/kube-apiserver

ARG CHART_TESTING_VERSION

RUN wget -O /tmp/ct-${CHART_TESTING_VERSION}-linux-amd64.tar.gz \
  https://github.com/helm/chart-testing/releases/download/v${CHART_TESTING_VERSION}/chart-testing_${CHART_TESTING_VERSION}_linux_amd64.tar.gz \
  && mkdir -p /tmp/ct-download \
  && tar xzvf /tmp/ct-${CHART_TESTING_VERSION}-linux-amd64.tar.gz -C /tmp/ct-download \
  && rm /tmp/ct-${CHART_TESTING_VERSION}-linux-amd64.tar.gz \
  && cp /tmp/ct-download/ct /usr/local/bin \
  && mkdir -p /etc/ct \
  && cp -R /tmp/ct-download/etc/* /etc/ct \
  && rm -rf /tmp/*

RUN wget -O /usr/local/bin/lj-releng \
  https://raw.githubusercontent.com/openresty/openresty-devel-utils/master/lj-releng \
  && chmod +x /usr/local/bin/lj-releng

ARG HELM_VERSION

RUN wget -O /tmp/helm.tgz \
  https://get.helm.sh/helm-${HELM_VERSION}-linux-${TARGETARCH}.tar.gz \
  && tar -C /tmp -xzf /tmp/helm.tgz \
  && cp /tmp/linux*/helm /usr/local/bin \
  && rm -rf /tmp/*

# Install a YAML Linter
ARG YAML_LINT_VERSION
RUN pip install "yamllint==$YAML_LINT_VERSION"

# Install Yamale YAML schema validator
ARG YAMALE_VERSION
RUN pip install "yamale==$YAMALE_VERSION"

WORKDIR $GOPATH
