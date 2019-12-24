# Copyright 2015 The Kubernetes Authors. All rights reserved.
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


FROM BASEIMAGE as builder

CROSS_BUILD_COPY qemu-ARCH-static /usr/bin/

RUN clean-install bash

COPY . /

RUN /build.sh

# Use a multi-stage build
FROM BASEIMAGE

ENV PATH=$PATH:/usr/local/luajit/bin:/usr/local/nginx/sbin:/usr/local/nginx/bin

ENV LUA_PATH="/usr/local/share/luajit-2.1.0-beta3/?.lua;/usr/local/share/lua/5.1/?.lua;/usr/local/lib/lua/?.lua;;"
ENV LUA_CPATH="/usr/local/lib/lua/?/?.so;/usr/local/lib/lua/?.so;;"

COPY --from=builder /usr/local /usr/local
COPY --from=builder /opt /opt
COPY --chown=www-data:www-data --from=builder /etc/nginx /etc/nginx

RUN apt-get update && apt-get dist-upgrade -y \
  && clean-install \
    bash \
    curl ca-certificates \
    libgeoip1 \
    patch \
    libpcre3 \
    zlib1g \
    libaio1 \
    openssl \
    util-linux \
    lmdb-utils \
    libcurl4 \
    libprotobuf17 \
    libz3-4 \
    procps \
    libxml2 libpcre++0v5 \
    liblmdb0 \
    libmaxminddb0 \
    dumb-init \
    nano \
    libyaml-cpp0.6 \
    libyajl2 \
  && ln -s /usr/local/nginx/sbin/nginx /sbin/nginx \
  && ln -s /usr/local/lib/mimalloc-1.2/libmimalloc.so /usr/local/lib/libmimalloc.so \
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

EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]
