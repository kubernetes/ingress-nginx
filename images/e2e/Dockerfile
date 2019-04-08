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

FROM quay.io/kubernetes-ingress-controller/nginx-amd64:0.84

RUN clean-install \
  g++ \
  gcc \
  git \
  libc6-dev \
  make \
  wget \
  luarocks \
  python \
  pkg-config

ENV GOLANG_VERSION 1.12.1
ENV GO_ARCH        linux-amd64
ENV GOLANG_SHA     2a3fdabf665496a0db5f41ec6af7a9b15a49fbe71a85a50ca38b1f13a103aeec

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

ENV RESTY_CLI_VERSION 0.23
ENV RESTY_CLI_SHA     8715b9a6e33140fd468779cd561c0c622f7444a902dc578b1137941ff3fc1ed8

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

RUN  curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.13.3/bin/linux/amd64/kubectl \
  && chmod +x /usr/local/bin/kubectl
