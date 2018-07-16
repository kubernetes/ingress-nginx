#!/usr/bin/dumb-init /bin/bash

# Copyright 2017 The Kubernetes Authors.
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

set -e

mkdir -p /var/log/nginx
echo 0 > /tmp/nginx.pid
writeDirs=( \
    /etc/nginx/template \
    /etc/ingress-controller/ssl \
    /etc/ingress-controller/auth \
    /var/log \
    /var/log/nginx \
    /tmp \
);

for dir in "${writeDirs[@]}"; do
    mkdir -p ${dir};
    chown -R www-data.www-data ${dir};
done

ln -sf /dev/stdout /var/log/nginx/access.log
ln -sf /dev/stderr /var/log/nginx/error.log

chown www-data.www-data /var/log/nginx/*
chown www-data.www-data /etc/nginx/nginx.conf
chown www-data.www-data /etc/nginx/opentracing.json
chown www-data.www-data /etc/nginx

echo "Testing if setcap is supported..."
if test 'setcap cap_net_bind_service=+ep /usr/sbin/nginx'; then
    echo "setcap is supported. Setting cap_net_bind_service=+ep to allow binding port lower than 1024 as non-root"
    setcap cap_net_bind_service=+ep    /usr/sbin/nginx
    setcap -v cap_net_bind_service=+ep /usr/sbin/nginx
    setcap cap_net_bind_service=+ep    /nginx-ingress-controller
    setcap -v cap_net_bind_service=+ep /nginx-ingress-controller

    echo "Droping root privileges and running as user..."
    su-exec www-data:www-data "$@"
else
    echo "WARNING!!!: setcap is not supported. Running as root"
    echo "Please check https://github.com/moby/moby/issues/1070"
    "$@"
fi
