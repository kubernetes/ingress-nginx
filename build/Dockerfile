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

FROM golang:1.10.3-stretch

RUN apt update \
  && apt install -y --no-install-recommends \
    luarocks \
  && apt-get clean -y \
  && rm -rf \
    /var/cache/debconf/* \
    /var/lib/apt/lists/* \
    /var/log/* \
    /tmp/* \
    /var/tmp/*

RUN luarocks install luacheck \
  && luarocks install busted 2.0.rc12 \
  && luarocks install lua-cjson 2.1.0-1

RUN go get github.com/onsi/ginkgo/ginkgo \
  && go get golang.org/x/lint/golint
