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
  dumb-init \
  libcap2-bin 

COPY . /

RUN setcap cap_net_bind_service=+ep /usr/sbin/nginx \
  && setcap cap_net_bind_service=+ep /nginx-ingress-controller

RUN bash -eux -c ' \
    writeDirs=( \
        /etc/nginx \
        /etc/ingress-controller/ssl \
        /etc/ingress-controller/auth \
        /var/log \
        /var/log/nginx \
        /opt/modsecurity/var/log \
        /opt/modsecurity/var/upload \
        /opt/modsecurity/var/audit \
    ); \
    for dir in "${writeDirs[@]}"; do \
      mkdir -p ${dir}; \
      chown -R www-data.www-data ${dir}; \
    done \
    '

# Create symlinks to redirect nginx logs to stdout and stderr docker log collector
# This only works if nginx is started with CMD or ENTRYPOINT
RUN  ln -sf /dev/stdout /var/log/nginx/access.log \
  && ln -sf /dev/stderr /var/log/nginx/error.log

USER www-data

ENTRYPOINT ["/usr/bin/dumb-init"]

CMD ["/nginx-ingress-controller"]
