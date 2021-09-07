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


FROM debian:bullseye-slim as builder

COPY . /

RUN apt-get update \
  && apt-get -y dist-upgrade \
  && /build.sh

# Use a multi-stage build
FROM debian:bullseye-slim

ENV PATH=$PATH:/usr/local/luajit/bin:/usr/local/nginx/sbin:/usr/local/nginx/bin

ENV LUA_PATH="/usr/local/share/luajit-2.1.0-beta3/?.lua;/usr/local/share/lua/5.1/?.lua;/usr/local/lib/lua/?.lua;;"
ENV LUA_CPATH="/usr/local/lib/lua/?/?.so;/usr/local/lib/lua/?.so;;"

ENV LD_LIBRARY_PATH="/usr/local/lib:/usr/local/modsecurity"

COPY --from=builder /usr/local /usr/local
COPY --from=builder /etc/nginx /etc/nginx

RUN apt-get update \
    && apt-get -y dist-upgrade \
    && apt-get -y --no-install-recommends install libgeoip1 \
    curl ca-certificates openssl dumb-init \
    tzdata \
    # Below libraries are part of modsecurity and need to be on the base image
    libxml2 libfuzzy2 libyajl2 liblmdb0 libmaxminddb0 \
    && ln -s /usr/local/nginx/sbin/nginx /sbin/nginx \
    && groupmod -g 101 www-data \
    && usermod -d /usr/local/nginx -u 101 www-data \
    && apt-get clean -y \
    && rm -rf \
       /var/cache/debconf/* \
       /var/lib/apt/lists/* \
       /var/log/* \
       /tmp/* \
       /var/tmp/* \
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
        chown -R www-data.www-data ${dir}; \
      done'
 
