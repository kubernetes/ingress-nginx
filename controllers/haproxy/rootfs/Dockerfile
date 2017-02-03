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

FROM haproxy:1.7-alpine
RUN apk --no-cache add openssl

# dumb-init kindly manages SIGCHLD from forked HAProxy processes
ARG DUMB_INIT_SHA256=81231da1cd074fdc81af62789fead8641ef3f24b6b07366a1c34e5b059faf363
RUN wget -O/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64\
 && echo "$DUMB_INIT_SHA256  /dumb-init" | sha256sum -c -\
 && chmod +x /dumb-init

COPY haproxy-ingress-controller /
COPY haproxy-wrapper /
COPY haproxy.tmpl /usr/local/etc/haproxy/

ENTRYPOINT ["/dumb-init", "--", "/haproxy-ingress-controller"]
