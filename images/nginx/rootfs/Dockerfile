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
FROM registry.access.redhat.com/ubi8/ubi:8.5 as builder

COPY . /

RUN /build.sh

# Use a multi-stage build
FROM registry.access.redhat.com/ubi8/ubi:8.5

ENV PATH=$PATH:/usr/local/luajit/bin:/usr/local/nginx/sbin:/usr/local/nginx/bin:/usr/local/lib:/lib64:/lib

ENV LUA_PATH="/usr/local/share/luajit-2.1.0-beta3/?.lua;/usr/local/share/lua/5.1/?.lua;/usr/local/lib/lua/?.lua;;"
ENV LUA_CPATH="/usr/local/lib/lua/?/?.so;/usr/local/lib/lua/?.so;;"
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

COPY --from=builder /usr/local /usr/local
COPY --from=builder /opt /opt
COPY --from=builder /etc/nginx /etc/nginx
COPY --from=builder entrypoint.sh /usr/local/entrypoint.sh

RUN touch /etc/ld.so.conf.d/local_libs.conf \
  && echo "/usr/local/lib" >> /etc/ld.so.conf.d/local_libs.conf \
  && echo "/usr/local/lib/lua" >> /etc/ld.so.conf.d/local_libs.conf \
  && echo "/usr/local/lib/mimalloc-1.6" >> /etc/ld.so.conf.d/local_libs.conf \
  && ldconfig -v

ARG TARGETARCH

RUN chmod +x /usr/local/entrypoint.sh \
  && yum -y update \
  && yum -y upgrade \
  && yum -y install \
    bash \
    pcre \
    zlib \
    curl ca-certificates \
    patch \
    yajl \
    libxml2 \
    libyaml \
    nano \
    openssl \
    tzdata \
    shadow-utils \
    wget \
  && yum clean all \
  && rm -rf \
    /var/cache/yum \
    /usr/share/doc \
  && ARCH=$(echo $TARGETARCH | sed -e s/^amd64$/x86_64/ -e s/^arm64$/aarch64/) \
  && wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.3/dumb-init_1.2.3_${ARCH} \
  && chmod +x /usr/bin/dumb-init \
  && ln -s /usr/local/nginx/sbin/nginx /sbin/nginx \
  && groupadd -rg 101 www-data \
  && adduser -u 101 -M -d /usr/local/nginx \
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
    chown -R www-data.www-data ${dir}; \
  done'

EXPOSE 80 443

ENTRYPOINT ["/usr/local/entrypoint.sh"]

CMD ["nginx", "-g", "daemon off;"]
