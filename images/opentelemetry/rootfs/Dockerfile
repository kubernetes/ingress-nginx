# Copyright 2021 The Kubernetes Authors. All rights reserved.
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


FROM alpine:3.20 AS base

RUN mkdir -p /opt/third_party/install
COPY . /opt/third_party/

# install build tools
RUN apk update \
	&& apk upgrade \
	&& apk add -U bash \
	&& bash /opt/third_party/build.sh -p

ENV NINJA_STATUS="[%p/%f/%t] "

# install otel_ngx_module.so
FROM base AS nginx
ARG NGINX_VERSION=1.25.3
RUN bash /opt/third_party/build.sh -n ${NGINX_VERSION}

FROM golang:1.22.8-bullseye AS build-init

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/init_module

FROM gcr.io/distroless/static-debian11 AS final
COPY --from=build-init /go/bin/init_module /
COPY --from=nginx /etc/nginx/modules /etc/nginx/modules

CMD ["/init_module"]
