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


FROM ubuntu:groovy-20200911 as builder

COPY . /

RUN /build.sh

# Use a multi-stage build
FROM ubuntu:groovy-20200911

ENV PATH=$PATH:/usr/local/luajit/bin:/usr/local/nginx/sbin:/usr/local/nginx/bin

ENV LUA_PATH="/usr/local/share/luajit-2.1.0-beta3/?.lua;/usr/local/share/lua/5.1/?.lua;/usr/local/lib/lua/?.lua;;"
ENV LUA_CPATH="/usr/local/lib/lua/?/?.so;/usr/local/lib/lua/?.so;;"

RUN userdel  www-data -f || true \
  && groupdel www-data -f || true \
  && addgroup --gid 101 www-data \
  && adduser  --gid 101 \
    --disabled-password --disabled-login \
    --gecos '' \
    --shell /usr/sbin/nologin \
    --uid 101 \
    --home /usr/local/nginx www-data

COPY --from=builder /usr/local /usr/local
COPY --from=builder /opt /opt
COPY --from=builder /etc/nginx /etc/nginx

RUN apt-get update && apt-get dist-upgrade -y \
  && apt-get install -y --no-install-recommends \
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
    libz3-4 \
    procps \
    libxml2 libpcre++0v5 \
    liblmdb0 \
    libmaxminddb0 \
    dumb-init \
    nano \
    libyaml-cpp0.6 \
    libyajl2 \
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

EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]
