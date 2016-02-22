#!/bin/sh

# Copyright 2015 The Kubernetes Authors All rights reserved.
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


set -eof pipefail

export NGINX_VERSION=1.9.11
export NDK_VERSION=0.2.19
export VTS_VERSION=0.1.8
export SETMISC_VERSION=0.29
export LUA_VERSION=0.10.1rc0 
export LUA_CJSON_VERSION=f79aa68af865ae84b36c7e794beedd87fef2ed54
export LUA_RESTY_HTTP_VERSION=0.06
export LUA_UPSTREAM_VERSION=0.04
export MORE_HEADERS_VERSION=0.29

export BUILD_PATH=/tmp/build

get_src()
{
  hash="$1"
  url="$2"
  f=$(basename "$url")

  curl -sSL "$url" -o "$f"
  echo "$hash  $f" | sha256sum -c - || exit 10
  tar xzf "$f"
  rm -rf "$f"
}

mkdir "$BUILD_PATH"
cd "$BUILD_PATH"

# install required packages to build
apk add --update-cache \
  bash \
  build-base \
  curl \
  geoip \
  geoip-dev \
  libcrypto1.0 \
  patch \
  pcre \
  pcre-dev \
  openssl-dev \
  zlib \
  zlib-dev \
  pcre-dev \
  libaio \
  libaio-dev \
  luajit \
  luajit-dev \
  linux-headers

# download, verify and extract the source files
get_src 6a5c72f4afaf57a6db064bba0965d72335f127481c5d4e64ee8714e7b368a51f \
        "http://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"

get_src 501f299abdb81b992a980bda182e5de5a4b2b3e275fbf72ee34dd7ae84c4b679 \
        "https://github.com/simpl/ngx_devel_kit/archive/v$NDK_VERSION.tar.gz"

get_src 8d280fc083420afb41dbe10df9a8ceec98f1d391bd2caa42ebae67d5bc9295d8 \
        "https://github.com/openresty/set-misc-nginx-module/archive/v$SETMISC_VERSION.tar.gz"

get_src 6bb9a36d8d70302d691c49557313fb7262cafd942a961d11a2730d9a5d9f70e0 \
        "https://github.com/vozlt/nginx-module-vts/archive/v$VTS_VERSION.tar.gz"

get_src 1bae94d2a0fd4fad39f2544a2f8eaf71335ea512a6f0027af190b46562224c68 \
        "https://github.com/openresty/lua-nginx-module/archive/v$LUA_VERSION.tar.gz"

get_src 2c451368a9e1a6fc01ed196cd6bd1602ee29f4b264df9263816e4dce17bca2c0 \
        "https://github.com/openresty/lua-cjson/archive/$LUA_CJSON_VERSION.tar.gz"

get_src 30ea2b03e8e8c4add5e143cc1826fd3364df58ea7f4b9a3fe02cd1630a505701 \
        "https://github.com/pintsized/lua-resty-http/archive/v$LUA_RESTY_HTTP_VERSION.tar.gz"

get_src 0a5f3003b5851373b03c542723eb5e7da44a01bf4c4c5f20b4de53f355a28d33 \
        "https://github.com/openresty/headers-more-nginx-module/archive/v$MORE_HEADERS_VERSION.tar.gz"

get_src eec4bbb40fd14e12179fd536a029e2fe82a7f29340ed357879d0b02b65302913 \
        "https://github.com/openresty/lua-upstream-nginx-module/archive/v$LUA_UPSTREAM_VERSION.tar.gz"

# build nginx
cd "$BUILD_PATH/nginx-$NGINX_VERSION"

./configure \
  --prefix=/usr \
  --conf-path=/etc/nginx/nginx.conf \
  --http-log-path=/var/log/nginx/access.log \
  --error-log-path=/var/log/nginx/error.log \
  --lock-path=/var/lock/nginx.lock \
  --pid-path=/run/nginx.pid \
  --http-client-body-temp-path=/var/lib/nginx/body \
  --http-fastcgi-temp-path=/var/lib/nginx/fastcgi \
  --http-proxy-temp-path=/var/lib/nginx/proxy \
  --http-scgi-temp-path=/var/lib/nginx/scgi \
  --http-uwsgi-temp-path=/var/lib/nginx/uwsgi \
  --with-debug \
  --with-pcre-jit \
  --with-ipv6 \
  --with-http_ssl_module \
  --with-http_stub_status_module \
  --with-http_realip_module \
  --with-http_auth_request_module \
  --with-http_addition_module \
  --with-http_dav_module \
  --with-http_geoip_module \
  --with-http_gzip_static_module \
  --with-http_sub_module \
  --with-http_v2_module \
  --with-stream \
  --with-stream_ssl_module \
  --with-threads \
  --with-file-aio \
  --add-module="$BUILD_PATH/ngx_devel_kit-$NDK_VERSION" \
  --add-module="$BUILD_PATH/set-misc-nginx-module-$SETMISC_VERSION" \
  --add-module="$BUILD_PATH/nginx-module-vts-$VTS_VERSION" \
  --add-module="$BUILD_PATH/lua-nginx-module-$LUA_VERSION" \
  --add-module="$BUILD_PATH/headers-more-nginx-module-$MORE_HEADERS_VERSION" \
  --add-module="$BUILD_PATH/lua-upstream-nginx-module-$LUA_UPSTREAM_VERSION" \
  && make && make install

echo "Installing CJSON module"
cd "$BUILD_PATH/lua-cjson-$LUA_CJSON_VERSION"
make LUA_INCLUDE_DIR=/usr/include/luajit-2.0 && make install

echo "Installing lua-resty-http module"
# copy lua module
cd "$BUILD_PATH/lua-resty-http-$LUA_RESTY_HTTP_VERSION"
sed -i 's/resty.http_headers/http_headers/' $BUILD_PATH/lua-resty-http-$LUA_RESTY_HTTP_VERSION/lib/resty/http.lua
cp $BUILD_PATH/lua-resty-http-$LUA_RESTY_HTTP_VERSION/lib/resty/http.lua /usr/local/lib/lua/5.1
cp $BUILD_PATH/lua-resty-http-$LUA_RESTY_HTTP_VERSION/lib/resty/http_headers.lua /usr/local/lib/lua/5.1

echo "Cleaning..."

cd /

rm -rf "$BUILD_PATH"

apk del --purge \
  build-base \
  geoip-dev \
  patch \
  openssl-dev \
  zlib-dev \
  pcre-dev \
  luajit-dev \
  libaio-dev \
  linux-headers

mkdir -p /var/lib/nginx/body /usr/share/nginx/html
mv /usr/html /usr/share/nginx

rm -rf /var/cache/apk/*
