#!/bin/bash

# Copyright 2015 The Kubernetes Authors.
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

export NGINX_VERSION=1.13.5
export NDK_VERSION=0.3.0
export VTS_VERSION=0.1.15
export SETMISC_VERSION=0.31
export STICKY_SESSIONS_VERSION=08a395c66e42
export MORE_HEADERS_VERSION=0.32
export NGINX_DIGEST_AUTH=7955af9c77598c697ac292811914ce1e2b3b824c
export NGINX_SUBSTITUTIONS=bc58cb11844bc42735bbaef7085ea86ace46d05b
export NGINX_OPENTRACING=fcc2e822c6dfc7d1f432c16b07dee9437c24236a
export OPENTRACING_CPP=42cbb358b68e53145c5b479efa09a25dbc81a95a
export ZIPKIN_CPP=8eae512bd750b304764d96058c65229f9a7712a9

export BUILD_PATH=/tmp/build

export NGINX_OPENTRACING_VENDOR="ZIPKIN"


ARCH=$(uname -p)

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

if [[ ${ARCH} == "ppc64le" ]]; then
  apt-get update && apt-get install --no-install-recommends -y software-properties-common
fi

# install required packages to build
apt-get update && apt-get install --no-install-recommends -y \
  bash \
  build-essential \
  curl ca-certificates \
  libgeoip1 \
  libgeoip-dev \
  patch \
  libpcre3 \
  libpcre3-dev \
  libssl-dev \
  zlib1g \
  zlib1g-dev \
  libaio1 \
  libaio-dev \
  openssl \
  libperl-dev \
  cmake \
  libcurl4-openssl-dev \
  linux-headers-generic || exit 1


# download, verify and extract the source files
get_src 0e75b94429b3f745377aeba3aff97da77bf2b03fcb9ff15b3bad9b038db29f2e \
        "http://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"

get_src 88e05a99a8a7419066f5ae75966fb1efc409bad4522d14986da074554ae61619 \
        "https://github.com/simpl/ngx_devel_kit/archive/v$NDK_VERSION.tar.gz"

get_src 97946a68937b50ab8637e1a90a13198fe376d801dc3e7447052e43c28e9ee7de \
        "https://github.com/openresty/set-misc-nginx-module/archive/v$SETMISC_VERSION.tar.gz"

get_src 5112a054b1b1edb4c0042a9a840ef45f22abb3c05c68174e28ebf483164fb7e1 \
        "https://github.com/vozlt/nginx-module-vts/archive/v$VTS_VERSION.tar.gz"

get_src c6d9dab8ea1fc997031007e2e8f47cced01417e203cd88d53a9fe9f6ae138720 \
        "https://github.com/openresty/headers-more-nginx-module/archive/v$MORE_HEADERS_VERSION.tar.gz"

get_src 53e440737ed1aff1f09fae150219a45f16add0c8d6e84546cb7d80f73ebffd90 \
        "https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng/get/$STICKY_SESSIONS_VERSION.tar.gz"

get_src 9b1d0075df787338bb607f14925886249bda60b6b3156713923d5d59e99a708b \
        "https://github.com/atomx/nginx-http-auth-digest/archive/$NGINX_DIGEST_AUTH.tar.gz"

get_src 618551948ab14cac51d6e4ad00452312c7b09938f59ebff4f93875013be31f2d \
        "https://github.com/yaoweibin/ngx_http_substitutions_filter_module/archive/$NGINX_SUBSTITUTIONS.tar.gz"

get_src d48c83e81aeeaebbf894adc5557cb8d027fb336d2afe95b68b2aa75920a3be74 \
        "https://github.com/rnburn/nginx-opentracing/archive/$NGINX_OPENTRACING.tar.gz"

get_src d1afc7c38bef055ac8a3759f117281b9d9287785e044a7d4e79134fa6ea99324 \
        "https://github.com/opentracing/opentracing-cpp/archive/$OPENTRACING_CPP.tar.gz"

get_src c9961b503da1119eaeb15393bac804a303d49d86f701dbfd31c425d2214354d5 \
        "https://github.com/rnburn/zipkin-cpp-opentracing/archive/$ZIPKIN_CPP.tar.gz"


#https://blog.cloudflare.com/optimizing-tls-over-tcp-to-reduce-latency/
curl -sSL -o nginx__dynamic_tls_records.patch https://raw.githubusercontent.com/cloudflare/sslconfig/master/patches/nginx__1.11.5_dynamic_tls_records.patch

# http2 header compression
curl -sSL -o nginx_http2_hpack.patch https://raw.githubusercontent.com/cloudflare/sslconfig/master/patches/nginx_http2_hpack.patch

# build opentracing lib
cd "$BUILD_PATH/opentracing-cpp-$OPENTRACING_CPP"
mkdir .build
cd .build
cmake ..
make
make install

# build zipkin lib
cd "$BUILD_PATH/zipkin-cpp-opentracing-$ZIPKIN_CPP"
mkdir .build
cd .build
cmake -DBUILD_SHARED_LIBS=1 ..
make
make install


# build nginx
cd "$BUILD_PATH/nginx-$NGINX_VERSION"

echo "Applying tls nginx patches..."
patch -p1 < $BUILD_PATH/nginx__dynamic_tls_records.patch
patch -p1 < $BUILD_PATH/nginx_http2_hpack.patch

WITH_FLAGS="--with-debug \
  --with-pcre-jit \
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
  --with-http_v2_hpack_enc \
  --with-stream \
  --with-stream_ssl_module \
  --with-stream_ssl_preread_module \
  --with-threads"

if [[ ${ARCH} != "armv7l" || ${ARCH} != "aarch64" ]]; then
  WITH_FLAGS+=" --with-file-aio"
fi

CC_OPT='-g -O3 -flto -fPIE -fstack-protector-strong -Wformat -Werror=format-security -Wdate-time -D_FORTIFY_SOURCE=2 --param=ssp-buffer-size=4 -DTCP_FASTOPEN=23 -Wno-error=strict-aliasing'
LD_OPT='-Wl,-Bsymbolic-functions -fPIE -pie -Wl,-z,relro -Wl,-z,now'

if [[ ${ARCH} == "x86_64" ]]; then
  CC_OPT+=' -m64 -mtune=generic'
fi

./configure \
  --prefix=/usr/share/nginx \
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
  ${WITH_FLAGS} \
  --without-mail_pop3_module \
  --without-mail_smtp_module \
  --without-mail_imap_module \
  --without-http_uwsgi_module \
  --without-http_scgi_module \
  --with-cc-opt="${CC_OPT}" \
  --with-ld-opt="${LD_OPT}" \
  --add-module="$BUILD_PATH/ngx_devel_kit-$NDK_VERSION" \
  --add-module="$BUILD_PATH/set-misc-nginx-module-$SETMISC_VERSION" \
  --add-module="$BUILD_PATH/nginx-module-vts-$VTS_VERSION" \
  --add-module="$BUILD_PATH/headers-more-nginx-module-$MORE_HEADERS_VERSION" \
  --add-module="$BUILD_PATH/nginx-goodies-nginx-sticky-module-ng-$STICKY_SESSIONS_VERSION" \
  --add-module="$BUILD_PATH/nginx-http-auth-digest-$NGINX_DIGEST_AUTH" \
  --add-module="$BUILD_PATH/ngx_http_substitutions_filter_module-$NGINX_SUBSTITUTIONS" \
  --add-module="$BUILD_PATH/nginx-opentracing-$NGINX_OPENTRACING" \
  && make || exit 1 \
  && make install || exit 1

echo "Cleaning..."

cd /

apt-mark unmarkauto \
  bash \
  curl ca-certificates \
  libgeoip1 \
  libpcre3 \
  zlib1g \
  libaio1 \
  xz-utils \
  geoip-bin \
  openssl

apt-get remove -y --purge \
  build-essential \
  gcc-5 \
  cpp-5 \
  libgeoip-dev \
  libpcre3-dev \
  libssl-dev \
  zlib1g-dev \
  libaio-dev \
  linux-libc-dev \
  perl-modules-5.22 \
  linux-headers-generic

apt-get autoremove -y

mkdir -p /var/lib/nginx/body /usr/share/nginx/html

mv /usr/share/nginx/sbin/nginx /usr/sbin

rm -rf "$BUILD_PATH"
rm -Rf /usr/share/man /usr/share/doc
rm -rf /tmp/* /var/tmp/*
rm -rf /var/lib/apt/lists/*
rm -rf /var/cache/apt/archives/*

# Download of GeoIP databases
curl -sSL -o /etc/nginx/GeoIP.dat.gz http://geolite.maxmind.com/download/geoip/database/GeoLiteCountry/GeoIP.dat.gz \
  && curl -sSL -o /etc/nginx/GeoLiteCity.dat.gz http://geolite.maxmind.com/download/geoip/database/GeoLiteCity.dat.gz \
  && gunzip /etc/nginx/GeoIP.dat.gz \
  && gunzip /etc/nginx/GeoLiteCity.dat.gz
