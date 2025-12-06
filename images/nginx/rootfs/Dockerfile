# Copyright 2024 The Kubernetes Authors. All rights reserved.
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
FROM alpine:3.23.0 AS builder

COPY . /

RUN apk update \
  && apk upgrade \
  && apk add -U bash --no-cache \
  && /build.sh

# Use a multi-stage build
FROM alpine:3.23.0

ENV PATH=$PATH:/usr/local/luajit/bin:/usr/local/nginx/sbin:/usr/local/nginx/bin

ENV LUA_PATH="/usr/local/share/luajit-2.1.0-beta3/?.lua;/usr/local/share/lua/5.1/?.lua;/usr/local/lib/lua/?.lua;;"
ENV LUA_CPATH="/usr/local/lib/lua/?/?.so;/usr/local/lib/lua/?.so;;"

COPY --from=builder /usr/local /usr/local
COPY --from=builder /usr/lib/libopentelemetry* /usr/local/lib
COPY --from=builder /opt /opt
COPY --from=builder /etc/nginx /etc/nginx

RUN apk update \
  && apk upgrade \
  && apk add -U --no-cache \
  bash \
  openssl \
  pcre \
  zlib \
  ca-certificates \
  patch \
  yajl \
  lmdb \
  libxml2 \
  libmaxminddb \
  yaml-cpp \
  dumb-init \
  tzdata \
  grpc-cpp \
  libprotobuf \
  abseil-cpp-crc-cpu-detect \
  abseil-cpp-vlog-config-internal \
  && ln -s /usr/local/nginx/sbin/nginx /sbin/nginx \
  && adduser -S -D -H -u 101 -h /usr/local/nginx \
  -s /sbin/nologin -G www-data -g www-data www-data \
  && bash -eu -c ' \
  writeDirs=( \
  /var/log/nginx \
  /var/lib/nginx/body \
  /var/lib/nginx/fastcgi \
  /var/lib/nginx/proxy \
  /var/lib/nginx/scgi \
  /var/lib/nginx/uwsgi \
  /var/log/audit \
  ); \
  for dir in "${writeDirs[@]}"; do \
  mkdir -p ${dir}; \
  chown -R www-data:www-data ${dir}; \
  done'

EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]
