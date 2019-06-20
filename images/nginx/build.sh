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


set -o errexit
set -o nounset
set -o pipefail

export NGINX_VERSION=1.13.7
export NDK_VERSION=0.3.0
export VTS_VERSION=0.1.15
export SETMISC_VERSION=0.31
export STICKY_SESSIONS_VERSION=08a395c66e42
export MORE_HEADERS_VERSION=0.33
export NGINX_DIGEST_AUTH=519dc2a4907bc6d9c48f95b3cf6a0151aaf44b40
export NGINX_SUBSTITUTIONS=bc58cb11844bc42735bbaef7085ea86ace46d05b
export NGINX_OPENTRACING_VERSION=0.1.1
export OPENTRACING_CPP_VERSION=1.0.0
export ZIPKIN_CPP_VERSION=0.1.0
export MODSECURITY=a2a5858d249222938c2f5e48087a922c63d7f9d8

export BUILD_PATH=/tmp/build

ARCH=$(uname -m)

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

if [[ ${ARCH} == "ppc64le" ]]; then
  clean-install software-properties-common
fi

# install required packages to build
clean-install \
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
  util-linux \
  wget \
  libcurl4-openssl-dev \
  procps \
  git g++ pkgconf flex bison doxygen libyajl-dev liblmdb-dev libgeoip-dev libtool dh-autoreconf libxml2 libpcre++-dev libxml2-dev \
  || exit 1

mkdir -p /etc/nginx

# download GeoIP databases
wget -O /etc/nginx/GeoIP.dat.gz https://geolite.maxmind.com/download/geoip/database/GeoLiteCountry/GeoIP.dat.gz || { echo 'Could not download GeoLiteCountry, exiting.' ; exit 1; }
wget -O /etc/nginx/GeoLiteCity.dat.gz https://geolite.maxmind.com/download/geoip/database/GeoLiteCity.dat.gz || { echo 'Could not download GeoLiteCity, exiting.' ; exit 1; }

gunzip /etc/nginx/GeoIP.dat.gz
gunzip /etc/nginx/GeoLiteCity.dat.gz

mkdir --verbose -p "$BUILD_PATH"
cd "$BUILD_PATH"

# download, verify and extract the source files
get_src beb732bc7da80948c43fd0bf94940a21a21b1c1ddfba0bd99a4b88e026220f5c \
        "http://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"

get_src 88e05a99a8a7419066f5ae75966fb1efc409bad4522d14986da074554ae61619 \
        "https://github.com/simpl/ngx_devel_kit/archive/v$NDK_VERSION.tar.gz"

get_src 97946a68937b50ab8637e1a90a13198fe376d801dc3e7447052e43c28e9ee7de \
        "https://github.com/openresty/set-misc-nginx-module/archive/v$SETMISC_VERSION.tar.gz"

get_src 5112a054b1b1edb4c0042a9a840ef45f22abb3c05c68174e28ebf483164fb7e1 \
        "https://github.com/vozlt/nginx-module-vts/archive/v$VTS_VERSION.tar.gz"

get_src a3dcbab117a9c103bc1ea5200fc00a7b7d2af97ff7fd525f16f8ac2632e30fbf \
        "https://github.com/openresty/headers-more-nginx-module/archive/v$MORE_HEADERS_VERSION.tar.gz"

get_src 53e440737ed1aff1f09fae150219a45f16add0c8d6e84546cb7d80f73ebffd90 \
        "https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng/get/$STICKY_SESSIONS_VERSION.tar.gz"

get_src 5ea7093cedea1a17b3c890015f83fc72346627400a9e6e3cec9c4b21297ecb61 \
        "https://github.com/atomx/nginx-http-auth-digest/archive/$NGINX_DIGEST_AUTH.tar.gz"

get_src 618551948ab14cac51d6e4ad00452312c7b09938f59ebff4f93875013be31f2d \
        "https://github.com/yaoweibin/ngx_http_substitutions_filter_module/archive/$NGINX_SUBSTITUTIONS.tar.gz"

get_src 8c55a7ec639b186689a0053eed5f1605f38c461f0d3ee6e728bb466a148a44d2 \
        "https://github.com/opentracing-contrib/nginx-opentracing/archive/v$NGINX_OPENTRACING_VERSION.tar.gz"

get_src 9543f66790ba65810869a29b3aaef5286f1c446cb498a304d0d8b153c289cae8 \
        "https://github.com/opentracing/opentracing-cpp/archive/v$OPENTRACING_CPP_VERSION.tar.gz"

get_src 3d91e866819986f5dda00549694c4eaca16f1209d1f87aecc8aca3b42aa72e08 \
        "https://github.com/rnburn/zipkin-cpp-opentracing/archive/v$ZIPKIN_CPP_VERSION.tar.gz"

get_src 3abdecedb5bf544eeba8c1ce0bef2da6a9f064b216ebbe20b68894afec1d7d80 \
        "https://github.com/SpiderLabs/ModSecurity-nginx/archive/$MODSECURITY.tar.gz"

#https://blog.cloudflare.com/optimizing-tls-over-tcp-to-reduce-latency/
curl -sSL -o nginx__dynamic_tls_records.patch https://raw.githubusercontent.com/cloudflare/sslconfig/master/patches/nginx__1.11.5_dynamic_tls_records.patch

# build opentracing lib
cd "$BUILD_PATH/opentracing-cpp-$OPENTRACING_CPP_VERSION"
mkdir .build
cd .build
cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=OFF ..
make
make install

# build zipkin lib
cd "$BUILD_PATH/zipkin-cpp-opentracing-$ZIPKIN_CPP_VERSION"
mkdir .build
cd .build
cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=1 -DBUILD_TESTING=OFF ..
make
make install

cd "$BUILD_PATH"

if [[ ${ARCH} != "s390x" ]]; then
 # Get Brotli source and deps
 git clone --depth=1 https://github.com/google/ngx_brotli.git 
 cd ngx_brotli && git submodule update --init
fi

if [[ ${ARCH} == "x86_64" ]]; then
  # build modsecurity library
  cd "$BUILD_PATH"
  git clone https://github.com/SpiderLabs/ModSecurity
  cd ModSecurity/
  git checkout -b v3/master origin/v3/master
  sh build.sh
  git submodule init
  git submodule update
  autoreconf -i
  automake
  autoconf
  ./configure
  make
  make install
fi

# build nginx
cd "$BUILD_PATH/nginx-$NGINX_VERSION"

echo "Applying nginx patches..."
patch -p1 < $BUILD_PATH/nginx__dynamic_tls_records.patch

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
  --with-stream \
  --with-stream_ssl_module \
  --with-stream_ssl_preread_module \
  --with-threads"

if [[ ${ARCH} != "armv7l" || ${ARCH} != "aarch64" ]]; then
  WITH_FLAGS+=" --with-file-aio"
fi

CC_OPT='-g -O3 -flto -fPIE -fstack-protector-strong -Wformat -Werror=format-security -Wdate-time -D_FORTIFY_SOURCE=2 --param=ssp-buffer-size=4 -DTCP_FASTOPEN=23 -Wno-error=strict-aliasing -fPIC'
LD_OPT='-Wl,-Bsymbolic-functions -fPIE -fPIC -pie -Wl,-z,relro -Wl,-z,now'

if [[ ${ARCH} == "x86_64" ]]; then
  CC_OPT+=' -m64 -mtune=generic'
fi

WITH_MODULES="--add-module=$BUILD_PATH/ngx_devel_kit-$NDK_VERSION \
  --add-module=$BUILD_PATH/set-misc-nginx-module-$SETMISC_VERSION \
  --add-module=$BUILD_PATH/nginx-module-vts-$VTS_VERSION \
  --add-module=$BUILD_PATH/headers-more-nginx-module-$MORE_HEADERS_VERSION \
  --add-module=$BUILD_PATH/nginx-goodies-nginx-sticky-module-ng-$STICKY_SESSIONS_VERSION \
  --add-module=$BUILD_PATH/nginx-http-auth-digest-$NGINX_DIGEST_AUTH \
  --add-module=$BUILD_PATH/ngx_http_substitutions_filter_module-$NGINX_SUBSTITUTIONS \
  --add-dynamic-module=$BUILD_PATH/nginx-opentracing-$NGINX_OPENTRACING_VERSION/opentracing \
  --add-dynamic-module=$BUILD_PATH/nginx-opentracing-$NGINX_OPENTRACING_VERSION/zipkin"

if [[ ${ARCH} == "x86_64" ]]; then
  WITH_MODULES+=" --add-dynamic-module=$BUILD_PATH/ModSecurity-nginx-$MODSECURITY"
fi

if [[ ${ARCH} != "s390x" ]]; then
  WITH_MODULES+=" --add-module=$BUILD_PATH/ngx_brotli"
fi

./configure \
  --prefix=/usr/share/nginx \
  --conf-path=/etc/nginx/nginx.conf \
  --modules-path=/etc/nginx/modules \
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
  ${WITH_MODULES} \
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
  libyajl2 liblmdb0 libxml2 libpcre++ \
  gzip \
  openssl

apt-get remove -y --purge \
  build-essential \
  gcc-6 \
  cpp-6 \
  libgeoip-dev \
  libpcre3-dev \
  libssl-dev \
  zlib1g-dev \
  libaio-dev \
  linux-libc-dev \
  cmake \
  wget \
  git g++ pkgconf flex bison doxygen libyajl-dev liblmdb-dev libgeoip-dev libtool dh-autoreconf libpcre++-dev libxml2-dev

apt-get autoremove -y

mkdir -p /var/lib/nginx/body /usr/share/nginx/html

mv /usr/share/nginx/sbin/nginx /usr/sbin

rm -rf "$BUILD_PATH"
rm -Rf /usr/share/man /usr/share/doc
rm -rf /tmp/* /var/tmp/*
rm -rf /var/lib/apt/lists/*
rm -rf /var/cache/apt/archives/*
rm -rf /usr/local/modsecurity/bin
rm -rf /usr/local/modsecurity/include
rm -rf /usr/local/modsecurity/lib/libmodsecurity.a

# Download owasp modsecurity crs
cd /etc/nginx/
curl -sSL -o master.tgz https://github.com/SpiderLabs/owasp-modsecurity-crs/archive/v3.0.2/master.tar.gz
tar zxpvf master.tgz
mv owasp-modsecurity-crs-3.0.2/ owasp-modsecurity-crs
rm master.tgz

cd owasp-modsecurity-crs
mv crs-setup.conf.example crs-setup.conf
mv rules/REQUEST-900-EXCLUSION-RULES-BEFORE-CRS.conf.example rules/REQUEST-900-EXCLUSION-RULES-BEFORE-CRS.conf
mv rules/RESPONSE-999-EXCLUSION-RULES-AFTER-CRS.conf.example rules/RESPONSE-999-EXCLUSION-RULES-AFTER-CRS.conf
cd ..

# Download modsecurity.conf
mkdir modsecurity
cd modsecurity
curl -sSL -o modsecurity.conf https://raw.githubusercontent.com/SpiderLabs/ModSecurity/v3/master/modsecurity.conf-recommended

# OWASP CRS v3 rules
echo "
Include /etc/nginx/owasp-modsecurity-crs/crs-setup.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-900-EXCLUSION-RULES-BEFORE-CRS.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-901-INITIALIZATION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-903.9001-DRUPAL-EXCLUSION-RULES.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-903.9002-WORDPRESS-EXCLUSION-RULES.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-905-COMMON-EXCEPTIONS.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-910-IP-REPUTATION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-911-METHOD-ENFORCEMENT.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-912-DOS-PROTECTION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-913-SCANNER-DETECTION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-920-PROTOCOL-ENFORCEMENT.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-921-PROTOCOL-ATTACK.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-930-APPLICATION-ATTACK-LFI.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-931-APPLICATION-ATTACK-RFI.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-932-APPLICATION-ATTACK-RCE.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-933-APPLICATION-ATTACK-PHP.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-942-APPLICATION-ATTACK-SQLI.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-943-APPLICATION-ATTACK-SESSION-FIXATION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-949-BLOCKING-EVALUATION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/RESPONSE-950-DATA-LEAKAGES.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/RESPONSE-951-DATA-LEAKAGES-SQL.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/RESPONSE-952-DATA-LEAKAGES-JAVA.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/RESPONSE-953-DATA-LEAKAGES-PHP.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/RESPONSE-954-DATA-LEAKAGES-IIS.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/RESPONSE-959-BLOCKING-EVALUATION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/RESPONSE-980-CORRELATION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/RESPONSE-999-EXCLUSION-RULES-AFTER-CRS.conf
" > /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf
