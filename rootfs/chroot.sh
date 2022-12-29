#!/bin/bash

# Copyright 2022 The Kubernetes Authors.
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

set -x
writeDirs=( \
  /chroot/etc/nginx \
  /chroot/usr/local/ \
  /chroot/etc/ingress-controller \
  /chroot/etc/ingress-controller/ssl \
  /chroot/etc/ingress-controller/auth \
  /chroot/opt/modsecurity/var/log \
  /chroot/opt/modsecurity/var/upload \
  /chroot/opt/modsecurity/var/audit \
  /chroot/var/log/audit \
  /chroot/var/lib/nginx \
  /chroot/var/log/nginx \
  /chroot/var/lib/nginx/body \
  /chroot/var/lib/nginx/fastcgi \
  /chroot/var/lib/nginx/proxy \
  /chroot/var/lib/nginx/scgi \
  /chroot/var/lib/nginx/uwsgi \
  /chroot/tmp/nginx
);

for dir in "${writeDirs[@]}"; do
  mkdir -p ${dir};
  chown -R www-data.www-data ${dir};
done

mkdir -p  /chroot/lib /chroot/proc /chroot/usr /chroot/bin /chroot/dev /chroot/run 
cp /etc/passwd /etc/group /chroot/etc/
cp -a /usr/* /chroot/usr/
cp -a /etc/nginx/* /chroot/etc/nginx/
cp /lib/ld-musl-* /lib/libcrypto* /lib/libssl* /lib/libz* /chroot/lib/
