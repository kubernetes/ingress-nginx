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

FROM BASEIMAGE

CROSS_BUILD_COPY qemu-QEMUARCH-static /usr/bin/

WORKDIR  /etc/nginx

RUN clean-install \
  diffutils \
  authbind netcat \
  dumb-init

COPY . /

# Fix permission during the build to avoid issues at runtime
# with volumes (custom templates)
RUN bash -eux -c ' \
  writeDirs=( \
    /etc/nginx/template \
    /etc/ingress-controller/ssl \
    /etc/ingress-controller/auth \
    /var/lib/nginx \
    /var/log/nginx \
    /tmp \
  ); \
  for dir in "${writeDirs[@]}"; do \
    mkdir -p ${dir}; \
    chown -R www-data.www-data ${dir}; \
  done' \
  && chmod 755 /etc/authbind/byuid/33 \
  && chown www-data /etc/authbind/byuid/33 \
  && chmod 755 /etc/authbind/byport/* \
  && chown www-data /etc/authbind/byport/* \
  && chown www-data /etc/nginx/opentracing.json

# Create symlinks to redirect nginx logs to stdout and stderr docker log collector
# This only works if nginx is started with CMD or ENTRYPOINT
RUN ln -sf /dev/stdout /var/lib/nginx/access.log
RUN ln -sf /dev/stderr /var/lib/nginx/error.log

ENTRYPOINT ["/usr/bin/dumb-init"]

CMD ["/nginx-ingress-controller"]
