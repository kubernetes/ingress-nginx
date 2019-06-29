# Copyright 2016 The Kubernetes Authors.
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

# This file creates a build environment for building and running kubernetes
# unit and integration tests

FROM gcr.io/k8s-testimages/bootstrap

# hint to kubetest that it is in CI
ENV KUBETEST_IN_DOCKER="true"

# Go standard envs
ENV GOPATH /go
ENV PATH /usr/local/go/bin:$PATH

RUN mkdir -p /go/bin
ENV PATH $GOPATH/bin:$PATH

# setup k8s repo symlink
RUN mkdir -p /go/src/k8s.io/kubernetes \
    && ln -s /go/src/k8s.io/kubernetes /workspace/kubernetes

# preinstall:
# - graphviz package for graphing profiles
# - bc for shell to junit
# - rpm for building RPMs with Bazel
RUN apt-get update \
    && apt-get install -y \
        bc \
        rpm \
    && rm -rf /var/lib/apt/lists/*

ARG K8S_RELEASE
ARG ETCD_VERSION

RUN curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl

RUN curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_RELEASE}/bin/linux/amd64/kube-apiserver -o /usr/local/bin/kube-apiserver \
    && chmod +x /usr/local/bin/kube-apiserver

RUN curl -sSL https://storage.googleapis.com/etcd/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VERSION}-linux-amd64.tar.gz \
    && mkdir -p /tmp/etcd-download \
    && tar xzvf /tmp/etcd-${ETCD_VERSION}-linux-amd64.tar.gz -C /tmp/etcd-download --strip-components=1 \
    && cp /tmp/etcd-download/etcd /usr/local/bin \
    && rm -rf /tmp/etcd-download

RUN curl -sSL https://github.com/kubernetes-sigs/kind/releases/download/v0.4.0/kind-linux-amd64 -o /usr/local/bin/kind \
    && chmod +x /usr/local/bin/kind

# install go
ENV GO_VERSION 1.12.6
ENV GO_TARBALL "go${GO_VERSION}.linux-amd64.tar.gz"
RUN wget -q "https://storage.googleapis.com/golang/${GO_TARBALL}" \
    && tar xzf "${GO_TARBALL}" -C /usr/local \
    && rm "${GO_TARBALL}"

RUN go get github.com/onsi/ginkgo/ginkgo \
    && go get golang.org/x/lint/golint \
    && rm -rf /go/src/github.com

ENV KUBEBUILDER_ASSETS /usr/local/bin
