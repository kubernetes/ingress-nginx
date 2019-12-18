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

FROM quay.io/kubernetes-ingress-controller/nginx-amd64:870be3bcd88c267f14fd82da82303472f383cd14

RUN clean-install \
  g++ \
  gcc \
  git \
  libc6-dev \
  make \
  wget \
  python \
  parallel \
  pkg-config

ENV GOLANG_VERSION 1.13.4
ENV GO_ARCH        linux-amd64
ENV GOLANG_SHA     692d17071736f74be04a72a06dab9cac1cd759377bd85316e52b2227604c004c

RUN set -eux; \
  url="https://golang.org/dl/go${GOLANG_VERSION}.${GO_ARCH}.tar.gz"; \
  wget -O go.tgz "$url"; \
  echo "${GOLANG_SHA} *go.tgz" | sha256sum -c -; \
  tar -C /usr/local -xzf go.tgz; \
  rm go.tgz; \
  export PATH="/usr/local/go/bin:$PATH"; \
  go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

WORKDIR $GOPATH

ENV RESTY_CLI_VERSION 0.25rc2
ENV RESTY_CLI_SHA     a38d850441384fa037a5922ca012dcce8708d0e4abe34ad2fe4164a01b28bdfb

RUN set -eux; \
  url="https://github.com/openresty/resty-cli/archive/v${RESTY_CLI_VERSION}.tar.gz"; \
  wget -O resty_cli.tgz "$url"; \
  echo "${RESTY_CLI_SHA} *resty_cli.tgz" | sha256sum -c -; \
  tar -C /tmp -xzf resty_cli.tgz; \
  rm resty_cli.tgz; \
  mv /tmp/resty-cli-${RESTY_CLI_VERSION}/bin/* /usr/local/bin/; \
  resty -V

RUN  luarocks install luacheck \
  && luarocks install busted

RUN  go get github.com/onsi/ginkgo/ginkgo \
  && go get golang.org/x/lint/golint

ARG K8S_RELEASE
ARG ETCD_VERSION

RUN wget https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl

RUN wget https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/amd64/kube-apiserver -O /usr/local/bin/kube-apiserver \
  && chmod +x /usr/local/bin/kube-apiserver

RUN curl -L https://storage.googleapis.com/etcd/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VERSION}-linux-amd64.tar.gz \
  && mkdir -p /tmp/etcd-download \
  && tar xzvf /tmp/etcd-${ETCD_VERSION}-linux-amd64.tar.gz -C /tmp/etcd-download --strip-components=1 \
  && cp /tmp/etcd-download/etcd /usr/local/bin \
  && rm -rf /tmp/etcd-download
