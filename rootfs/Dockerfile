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
  libcap2-bin

COPY --chown=www-data:www-data . /

RUN cp /usr/local/openresty/nginx/conf/mime.types /etc/nginx/mime.types
RUN ln -s /usr/local/openresty/nginx/modules /etc/nginx/modules

# Add LuaRocks paths
ENV LUA_PATH="/usr/local/openresty/site/lualib/?.ljbc;/usr/local/openresty/site/lualib/?/init.ljbc;/usr/local/openresty/lualib/?.ljbc;/usr/local/openresty/lualib/?/init.ljbc;/usr/local/openresty/site/lualib/?.lua;/usr/local/openresty/site/lualib/?/init.lua;/usr/local/openresty/lualib/?.lua;/usr/local/openresty/lualib/?/init.lua;./?.lua;/usr/local/openresty/luajit/share/luajit-2.1.0-beta3/?.lua;/usr/local/share/lua/5.1/?.lua;/usr/local/share/lua/5.1/?/init.lua;/usr/local/openresty/luajit/share/lua/5.1/?.lua;/usr/local/openresty/luajit/share/lua/5.1/?/init.lua;/usr/local/lib/lua/?.lua;/etc/nginx/lua/?.lua;/etc/nginx/lua/vendor/?.lua;;"
ENV LUA_CPATH="/etc/nginx/lua/?.lua;/etc/nginx/lua/vendor/?.lua;/usr/local/openresty/site/lualib/?.so;/usr/local/openresty/lualib/?.so;./?.so;/usr/local/lib/lua/5.1/?.so;/usr/local/openresty/luajit/lib/lua/5.1/?.so;/usr/local/lib/lua/5.1/loadall.so;/usr/local/openresty/luajit/lib/lua/5.1/?.so;;"

# Fix permission during the build to avoid issues at runtime
# with volumes (custom templates)
RUN bash -eu -c ' \
  writeDirs=( \
    /etc/ingress-controller/ssl \
    /etc/ingress-controller/auth \
    /var/log \
    /var/log/nginx \
    /tmp \
  ); \
  for dir in "${writeDirs[@]}"; do \
    mkdir -p ${dir}; \
    chown -R www-data.www-data ${dir}; \
  done'

RUN  setcap    cap_net_bind_service=+ep /nginx-ingress-controller \
  && setcap -v cap_net_bind_service=+ep /nginx-ingress-controller

RUN  setcap    cap_net_bind_service=+ep /usr/local/openresty/nginx/sbin/nginx \
  && setcap -v cap_net_bind_service=+ep /usr/local/openresty/nginx/sbin/nginx

USER www-data

# Create symlinks to redirect nginx logs to stdout and stderr docker log collector
RUN  ln -sf /dev/stdout /usr/local/openresty/nginx/logs/access.log \
  && ln -sf /dev/stderr /usr/local/openresty/nginx/logs/error.log \
  && ln -s /usr/local/openresty/nginx/logs/* /var/log/nginx

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/nginx-ingress-controller"]
