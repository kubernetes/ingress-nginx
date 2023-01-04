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

export NGINX_VERSION=1.21.6

# Check for recent changes: https://github.com/vision5/ngx_devel_kit/compare/v0.3.1...master
export NDK_VERSION=0.3.1

# Check for recent changes: https://github.com/openresty/set-misc-nginx-module/compare/v0.33...master
export SETMISC_VERSION=0.33

# Check for recent changes: https://github.com/openresty/headers-more-nginx-module/compare/v0.33...master
export MORE_HEADERS_VERSION=0.33

# Check for recent changes: https://github.com/atomx/nginx-http-auth-digest/compare/v1.0.0...atomx:master
export NGINX_DIGEST_AUTH=1.0.0

# Check for recent changes: https://github.com/yaoweibin/ngx_http_substitutions_filter_module/compare/v0.6.4...master
export NGINX_SUBSTITUTIONS=b8a71eacc7f986ba091282ab8b1bbbc6ae1807e0

# Check for recent changes: https://github.com/opentracing-contrib/nginx-opentracing/compare/v0.19.0...master
export NGINX_OPENTRACING_VERSION=0.19.0

#Check for recent changes: https://github.com/opentracing/opentracing-cpp/compare/v1.6.0...master
export OPENTRACING_CPP_VERSION=f86b33f3d9e7322b1298ba62d5ffa7a9519c4c41

# Check for recent changes: https://github.com/rnburn/zipkin-cpp-opentracing/compare/v0.5.2...master
export ZIPKIN_CPP_VERSION=f69593138ff84ca2f6bc115992e18ca3d35f344a

# Check for recent changes: https://github.com/jbeder/yaml-cpp/compare/yaml-cpp-0.7.0...master
export YAML_CPP_VERSION=yaml-cpp-0.7.0

# Check for recent changes: https://github.com/jaegertracing/jaeger-client-cpp/compare/v0.7.0...master
export JAEGER_VERSION=0.7.0

# Check for recent changes: https://github.com/msgpack/msgpack-c/compare/cpp-3.3.0...master
export MSGPACK_VERSION=3.3.0

# Check for recent changes: https://github.com/DataDog/dd-opentracing-cpp/compare/v1.3.2...master
export DATADOG_CPP_VERSION=1.3.2

# Check for recent changes: https://github.com/SpiderLabs/ModSecurity-nginx/compare/v1.0.3...master
export MODSECURITY_VERSION=1.0.3

# Check for recent changes: https://github.com/SpiderLabs/ModSecurity/compare/v3.0.8...v3/master
export MODSECURITY_LIB_VERSION=e9a7ba4a60be48f761e0328c6dfcc668d70e35a0

# Check for recent changes: https://github.com/coreruleset/coreruleset/compare/v3.3.2...v3.3/master
export OWASP_MODSECURITY_CRS_VERSION=v3.3.4

# Check for recent changes: https://github.com/openresty/lua-nginx-module/compare/v0.10.21...master
export LUA_NGX_VERSION=0.10.21

# Check for recent changes: https://github.com/openresty/stream-lua-nginx-module/compare/v0.0.11...master
export LUA_STREAM_NGX_VERSION=0.0.11

# Check for recent changes: https://github.com/openresty/lua-upstream-nginx-module/compare/8aa93ead98ba2060d4efd594ae33a35d153589bf...master
export LUA_UPSTREAM_VERSION=8aa93ead98ba2060d4efd594ae33a35d153589bf

# Check for recent changes: https://github.com/openresty/lua-cjson/compare/2.1.0.10...openresty:master
export LUA_CJSON_VERSION=2.1.0.10

export NGINX_INFLUXDB_VERSION=5b09391cb7b9a889687c0aa67964c06a2d933e8b

# Check for recent changes: https://github.com/leev/ngx_http_geoip2_module/compare/3.3...master
export GEOIP2_VERSION=a26c6beed77e81553686852dceb6c7fdacc5970d

# Check for recent changes: https://github.com/msva/nginx_ajp_module/compare/fcbb2ccca4901d317ecd7a9dabb3fec9378ff40f...master
# This is a fork from https://github.com/yaoweibin/nginx_ajp_module
# Since it has not been updated and is not compatible with NGINX 1.21
export NGINX_AJP_VERSION=fcbb2ccca4901d317ecd7a9dabb3fec9378ff40f

# Check for recent changes: https://github.com/openresty/luajit2/compare/v2.1-20220411...v2.1-agentzh
export LUAJIT_VERSION=2.1-20220411

# Check for recent changes: https://github.com/openresty/lua-resty-balancer/compare/v0.04...master
export LUA_RESTY_BALANCER=0.04

# Check for recent changes: https://github.com/openresty/lua-resty-lrucache/compare/v0.11...master
export LUA_RESTY_CACHE=0.11

# Check for recent changes: https://github.com/openresty/lua-resty-core/compare/v0.1.23...master
export LUA_RESTY_CORE=0.1.23

# Check for recent changes: https://github.com/cloudflare/lua-resty-cookie/compare/v0.1.0...master
export LUA_RESTY_COOKIE_VERSION=303e32e512defced053a6484bc0745cf9dc0d39e

# Check for recent changes: https://github.com/openresty/lua-resty-dns/compare/v0.22...master
export LUA_RESTY_DNS=0.22

# Check for recent changes: https://github.com/ledgetech/lua-resty-http/compare/v0.16.1...master
export LUA_RESTY_HTTP=0ce55d6d15da140ecc5966fa848204c6fd9074e8

# Check for recent changes: https://github.com/openresty/lua-resty-lock/compare/v0.08...master
export LUA_RESTY_LOCK=0.08

# Check for recent changes: https://github.com/openresty/lua-resty-upload/compare/v0.10...master
export LUA_RESTY_UPLOAD_VERSION=0.10

# Check for recent changes: https://github.com/openresty/lua-resty-string/compare/v0.15...master
export LUA_RESTY_STRING_VERSION=0.15

# Check for recent changes: https://github.com/openresty/lua-resty-memcached/compare/v0.16...master
export LUA_RESTY_MEMCACHED_VERSION=0.16

# Check for recent changes: https://github.com/openresty/lua-resty-redis/compare/v0.30...master
export LUA_RESTY_REDIS_VERSION=0.30

# Check for recent changes: https://github.com/api7/lua-resty-ipmatcher/compare/v0.6.1...master
export LUA_RESTY_IPMATCHER_VERSION=0.6.1

# Check for recent changes: https://github.com/ElvinEfendi/lua-resty-global-throttle/compare/v0.2.0...main
export LUA_RESTY_GLOBAL_THROTTLE_VERSION=0.2.0

# Check for recent changes:  https://github.com/microsoft/mimalloc/compare/v1.7.6...master
export MIMALOC_VERSION=1.7.6

export BUILD_PATH=/tmp/build

ARCH=$(uname -m)

if [[ ${ARCH} == "s390x" ]]; then
  export LUAJIT_VERSION=9d5750d28478abfdcaefdfdc408f87752a21e431
  export LUA_RESTY_CORE=0.1.17
  export LUA_NGX_VERSION=0.10.15
  export LUA_STREAM_NGX_VERSION=0.0.7
fi

get_src()
{
  hash="$1"
  url="$2"
  f=$(basename "$url")

  echo "Downloading $url"

  curl -sSL "$url" -o "$f"
  echo "$hash  $f" | sha256sum -c - || exit 10
  tar xzf "$f"
  rm -rf "$f"
}

# install required packages to build
apk add \
  bash \
  gcc \
  clang \
  libc-dev \
  make \
  automake \
  openssl-dev \
  pcre-dev \
  zlib-dev \
  linux-headers \
  libxslt-dev \
  gd-dev \
  geoip-dev \
  perl-dev \
  libedit-dev \
  mercurial \
  alpine-sdk \
  findutils \
  curl \
  ca-certificates \
  patch \
  libaio-dev \
  openssl \
  cmake \
  util-linux \
  lmdb-tools \
  wget \
  curl-dev \
  libprotobuf \
  git g++ pkgconf flex bison doxygen yajl-dev lmdb-dev libtool autoconf libxml2 libxml2-dev \
  python3 \
  libmaxminddb-dev \
  bc \
  unzip \
  dos2unix \
  yaml-cpp \
  coreutils

mkdir -p /etc/nginx

mkdir --verbose -p "$BUILD_PATH"
cd "$BUILD_PATH"

# download, verify and extract the source files
get_src 66dc7081488811e9f925719e34d1b4504c2801c81dee2920e5452a86b11405ae \
        "https://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"

get_src 0e971105e210d272a497567fa2e2c256f4e39b845a5ba80d373e26ba1abfbd85 \
        "https://github.com/simpl/ngx_devel_kit/archive/v$NDK_VERSION.tar.gz"

get_src cd5e2cc834bcfa30149e7511f2b5a2183baf0b70dc091af717a89a64e44a2985 \
        "https://github.com/openresty/set-misc-nginx-module/archive/v$SETMISC_VERSION.tar.gz"

get_src a3dcbab117a9c103bc1ea5200fc00a7b7d2af97ff7fd525f16f8ac2632e30fbf \
        "https://github.com/openresty/headers-more-nginx-module/archive/v$MORE_HEADERS_VERSION.tar.gz"

get_src f09851e6309560a8ff3e901548405066c83f1f6ff88aa7171e0763bd9514762b \
        "https://github.com/atomx/nginx-http-auth-digest/archive/v$NGINX_DIGEST_AUTH.tar.gz"

get_src a98b48947359166326d58700ccdc27256d2648218072da138ab6b47de47fbd8f \
        "https://github.com/yaoweibin/ngx_http_substitutions_filter_module/archive/$NGINX_SUBSTITUTIONS.tar.gz"

get_src 6f97776ebdf019b105a755c7736b70bdbd7e575c7f0d39db5fe127873c7abf17 \
        "https://github.com/opentracing-contrib/nginx-opentracing/archive/v$NGINX_OPENTRACING_VERSION.tar.gz"

get_src cbe625cba85291712253db5bc3870d60c709acfad9a8af5a302673d3d201e3ea \
        "https://github.com/opentracing/opentracing-cpp/archive/$OPENTRACING_CPP_VERSION.tar.gz"

get_src 71de3d0658935db7ccea20e006b35e58ddc7e4c18878b9523f2addc2371e9270 \
        "https://github.com/rnburn/zipkin-cpp-opentracing/archive/$ZIPKIN_CPP_VERSION.tar.gz"

get_src 32a42256616cc674dca24c8654397390adff15b888b77eb74e0687f023c8751b \
        "https://github.com/SpiderLabs/ModSecurity-nginx/archive/v$MODSECURITY_VERSION.tar.gz"

get_src 43e6a9fcb146ad871515f0d0873947e5d497a1c9c60c58cb102a97b47208b7c3 \
        "https://github.com/jbeder/yaml-cpp/archive/$YAML_CPP_VERSION.tar.gz"

get_src 3a3a03060bf5e3fef52c9a2de02e6035cb557f389453d8f3b0c1d3d570636994 \
        "https://github.com/jaegertracing/jaeger-client-cpp/archive/v$JAEGER_VERSION.tar.gz"

get_src 754c3ace499a63e45b77ef4bcab4ee602c2c414f58403bce826b76ffc2f77d0b \
        "https://github.com/msgpack/msgpack-c/archive/cpp-$MSGPACK_VERSION.tar.gz"

if [[ ${ARCH} == "s390x" ]]; then
get_src 7d5f3439c8df56046d0564b5857fd8a30296ab1bd6df0f048aed7afb56a0a4c2 \
        "https://github.com/openresty/lua-nginx-module/archive/v$LUA_NGX_VERSION.tar.gz"
get_src 99c47c75c159795c9faf76bbb9fa58e5a50b75286c86565ffcec8514b1c74bf9 \
        "https://github.com/openresty/stream-lua-nginx-module/archive/v$LUA_STREAM_NGX_VERSION.tar.gz"
else
get_src 9db756000578efaecb43bea4fc6cf631aaa80988d86ffe5d3afeb9927895ffad \
        "https://github.com/openresty/lua-nginx-module/archive/v$LUA_NGX_VERSION.tar.gz"

get_src c7924f28cb014a99636e747ea907724dd55f60e180cb92cde6e8ed48d2278f27 \
        "https://github.com/openresty/stream-lua-nginx-module/archive/v$LUA_STREAM_NGX_VERSION.tar.gz"

fi

get_src a92c9ee6682567605ece55d4eed5d1d54446ba6fba748cff0a2482aea5713d5f \
        "https://github.com/openresty/lua-upstream-nginx-module/archive/$LUA_UPSTREAM_VERSION.tar.gz"

if [[ ${ARCH} == "s390x" ]]; then
get_src 266ed1abb70a9806d97cb958537a44b67db6afb33d3b32292a2d68a2acedea75 \
        "https://github.com/openresty/luajit2/archive/$LUAJIT_VERSION.tar.gz"
else
get_src d3f2c870f8f88477b01726b32accab30f6e5d57ae59c5ec87374ff73d0794316 \
        "https://github.com/openresty/luajit2/archive/v$LUAJIT_VERSION.tar.gz"
fi

get_src 586f92166018cc27080d34e17c59d68219b85af745edf3cc9fe41403fc9b4ac6 \
        "https://github.com/DataDog/dd-opentracing-cpp/archive/v$DATADOG_CPP_VERSION.tar.gz"

get_src 1af5a5632dc8b00ae103d51b7bf225de3a7f0df82f5c6a401996c080106e600e \
        "https://github.com/influxdata/nginx-influxdb-module/archive/$NGINX_INFLUXDB_VERSION.tar.gz"

get_src 4c1933434572226942c65b2f2b26c8a536ab76aa771a3c7f6c2629faa764976b \
        "https://github.com/leev/ngx_http_geoip2_module/archive/$GEOIP2_VERSION.tar.gz"

get_src 778fcca851bd69dabfb382dc827d2ee07662f7eca36b5e66e67d5512bad75ef8 \
        "https://github.com/msva/nginx_ajp_module/archive/$NGINX_AJP_VERSION.tar.gz"

get_src 5d16e623d17d4f42cc64ea9cfb69ca960d313e12f5d828f785dd227cc483fcbd \
        "https://github.com/openresty/lua-resty-upload/archive/v$LUA_RESTY_UPLOAD_VERSION.tar.gz"

get_src bdbf271003d95aa91cab0a92f24dca129e99b33f79c13ebfcdbbcbb558129491 \
        "https://github.com/openresty/lua-resty-string/archive/v$LUA_RESTY_STRING_VERSION.tar.gz"

get_src 16d72ed133f0c6df376a327386c3ef4e9406cf51003a700737c3805770ade7c5 \
        "https://github.com/openresty/lua-resty-balancer/archive/v$LUA_RESTY_BALANCER.tar.gz"

if [[ ${ARCH} == "s390x" ]]; then
get_src 8f5f76d2689a3f6b0782f0a009c56a65e4c7a4382be86422c9b3549fe95b0dc4 \
        "https://github.com/openresty/lua-resty-core/archive/v$LUA_RESTY_CORE.tar.gz"
else
get_src efd6b51520429e64b1bcc10f477d370ebed1631c190f7e4dc270d959a743ad7d \
        "https://github.com/openresty/lua-resty-core/archive/v$LUA_RESTY_CORE.tar.gz"
fi

get_src 0c551d6898f89f876e48730f9b55790d0ba07d5bc0aa6c76153277f63c19489f \
        "https://github.com/openresty/lua-cjson/archive/$LUA_CJSON_VERSION.tar.gz"

get_src 5ed48c36231e2622b001308622d46a0077525ac2f751e8cc0c9905914254baa4 \
        "https://github.com/cloudflare/lua-resty-cookie/archive/$LUA_RESTY_COOKIE_VERSION.tar.gz"

get_src e810ed124fe788b8e4aac2c8960dda1b9a6f8d0ca94ce162f28d3f4d877df8af \
        "https://github.com/openresty/lua-resty-lrucache/archive/v$LUA_RESTY_CACHE.tar.gz"

get_src 2b4683f9abe73e18ca00345c65010c9056777970907a311d6e1699f753141de2 \
        "https://github.com/openresty/lua-resty-lock/archive/v$LUA_RESTY_LOCK.tar.gz"

get_src 70e9a01eb32ccade0d5116a25bcffde0445b94ad35035ce06b94ccd260ad1bf0 \
        "https://github.com/openresty/lua-resty-dns/archive/v$LUA_RESTY_DNS.tar.gz"

get_src 9fcb6db95bc37b6fce77d3b3dc740d593f9d90dce0369b405eb04844d56ac43f \
        "https://github.com/ledgetech/lua-resty-http/archive/$LUA_RESTY_HTTP.tar.gz"

get_src 42893da0e3de4ec180c9bf02f82608d78787290a70c5644b538f29d243147396 \
        "https://github.com/openresty/lua-resty-memcached/archive/v$LUA_RESTY_MEMCACHED_VERSION.tar.gz"

get_src c15aed1a01c88a3a6387d9af67a957dff670357f5fdb4ee182beb44635eef3f1 \
        "https://github.com/openresty/lua-resty-redis/archive/v$LUA_RESTY_REDIS_VERSION.tar.gz"

get_src efb767487ea3f6031577b9b224467ddbda2ad51a41c5867a47582d4ad85d609e \
        "https://github.com/api7/lua-resty-ipmatcher/archive/v$LUA_RESTY_IPMATCHER_VERSION.tar.gz"

get_src 0fb790e394510e73fdba1492e576aaec0b8ee9ef08e3e821ce253a07719cf7ea \
        "https://github.com/ElvinEfendi/lua-resty-global-throttle/archive/v$LUA_RESTY_GLOBAL_THROTTLE_VERSION.tar.gz"

get_src d74f86ada2329016068bc5a243268f1f555edd620b6a7d6ce89295e7d6cf18da \
        "https://github.com/microsoft/mimalloc/archive/refs/tags/v${MIMALOC_VERSION}.tar.gz"

# improve compilation times
CORES=$(($(grep -c ^processor /proc/cpuinfo) - 1))

export MAKEFLAGS=-j${CORES}
export CTEST_BUILD_FLAGS=${MAKEFLAGS}
export HUNTER_JOBS_NUMBER=${CORES}
export HUNTER_USE_CACHE_SERVERS=true

# Install luajit from openresty fork
export LUAJIT_LIB=/usr/local/lib
export LUA_LIB_DIR="$LUAJIT_LIB/lua"
export LUAJIT_INC=/usr/local/include/luajit-2.1

cd "$BUILD_PATH/luajit2-$LUAJIT_VERSION"
make CCDEBUG=-g
make install

ln -s /usr/local/bin/luajit /usr/local/bin/lua
ln -s "$LUAJIT_INC" /usr/local/include/lua

cd "$BUILD_PATH"

# Git tuning
git config --global --add core.compression -1

# build opentracing lib
cd "$BUILD_PATH/opentracing-cpp-$OPENTRACING_CPP_VERSION"
mkdir .build
cd .build

cmake -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_TESTING=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DBUILD_MOCKTRACER=OFF \
      -DBUILD_STATIC_LIBS=ON \
      -DCMAKE_POSITION_INDEPENDENT_CODE:BOOL=true \
      ..

make
make install

# build yaml-cpp
# TODO @timmysilv: remove this and jaeger sed calls once it is fixed in jaeger-client-cpp
cd "$BUILD_PATH/yaml-cpp-$YAML_CPP_VERSION"
mkdir .build
cd .build

cmake -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_POSITION_INDEPENDENT_CODE:BOOL=true \
      -DYAML_BUILD_SHARED_LIBS=ON \
      -DYAML_CPP_BUILD_TESTS=OFF \
      -DYAML_CPP_BUILD_TOOLS=OFF \
      ..

make
make install

# build jaeger lib
cd "$BUILD_PATH/jaeger-client-cpp-$JAEGER_VERSION"
sed -i 's/-Werror/-Wno-psabi/' CMakeLists.txt
# use the above built yaml-cpp instead until a new version of jaeger-client-cpp fixes the yaml-cpp issue
# tl;dr new hunter is needed for new yaml-cpp, but new hunter has a conflict with old Thrift and new Boost
sed -i 's/hunter_add_package(yaml-cpp)/#hunter_add_package(yaml-cpp)/' CMakeLists.txt
sed -i 's/yaml-cpp::yaml-cpp/yaml-cpp/' CMakeLists.txt

cat <<EOF > export.map
{
    global:
        OpenTracingMakeTracerFactory;
    local: *;
};
EOF

mkdir .build
cd .build

cmake -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_TESTING=OFF \
      -DJAEGERTRACING_BUILD_EXAMPLES=OFF \
      -DJAEGERTRACING_BUILD_CROSSDOCK=OFF \
      -DJAEGERTRACING_COVERAGE=OFF \
      -DJAEGERTRACING_PLUGIN=ON \
      -DHUNTER_CONFIGURATION_TYPES=Release \
      -DBUILD_SHARED_LIBS=OFF \
      -DJAEGERTRACING_WITH_YAML_CPP=ON \
      -DCMAKE_POSITION_INDEPENDENT_CODE:BOOL=true \
      ..

make
make install

export HUNTER_INSTALL_DIR=$(cat _3rdParty/Hunter/install-root-dir) \

mv libjaegertracing_plugin.so /usr/local/lib/libjaegertracing_plugin.so


# build zipkin lib
cd "$BUILD_PATH/zipkin-cpp-opentracing-$ZIPKIN_CPP_VERSION"

cat <<EOF > export.map
{
    global:
        OpenTracingMakeTracerFactory;
    local: *;
};
EOF

mkdir .build
cd .build

cmake -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_SHARED_LIBS=OFF \
      -DBUILD_PLUGIN=ON \
      -DBUILD_TESTING=OFF \
      -DCMAKE_POSITION_INDEPENDENT_CODE:BOOL=true \
      ..

make
make install

# build msgpack lib
cd "$BUILD_PATH/msgpack-c-cpp-$MSGPACK_VERSION"

mkdir .build
cd .build
cmake -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_SHARED_LIBS=OFF \
      -DMSGPACK_BUILD_EXAMPLES=OFF \
      -DCMAKE_POSITION_INDEPENDENT_CODE:BOOL=true \
      ..

make
make install

# build datadog lib
cd "$BUILD_PATH/dd-opentracing-cpp-$DATADOG_CPP_VERSION"

mkdir .build
cd .build

cmake -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_TESTING=OFF \
      -DCMAKE_POSITION_INDEPENDENT_CODE:BOOL=true \
      ..

make
make install

# Get Brotli source and deps
cd "$BUILD_PATH"
git clone --depth=1 https://github.com/google/ngx_brotli.git
cd ngx_brotli
git submodule init
git submodule update

cd "$BUILD_PATH"
git clone --depth=1 https://github.com/ssdeep-project/ssdeep
cd ssdeep/

./bootstrap
./configure

make
make install

# build modsecurity library
cd "$BUILD_PATH"
git clone -n https://github.com/SpiderLabs/ModSecurity
cd ModSecurity/
git checkout $MODSECURITY_LIB_VERSION
git submodule init
git submodule update

sh build.sh

# https://github.com/SpiderLabs/ModSecurity/issues/1909#issuecomment-465926762
sed -i '115i LUA_CFLAGS="${LUA_CFLAGS} -DWITH_LUA_JIT_2_1"' build/lua.m4
sed -i '117i AC_SUBST(LUA_CFLAGS)' build/lua.m4

./configure \
  --disable-doxygen-doc \
  --disable-doxygen-html \
  --disable-examples

make
make install

mkdir -p /etc/nginx/modsecurity
cp modsecurity.conf-recommended /etc/nginx/modsecurity/modsecurity.conf
cp unicode.mapping /etc/nginx/modsecurity/unicode.mapping

# Replace serial logging with concurrent
sed -i 's|SecAuditLogType Serial|SecAuditLogType Concurrent|g' /etc/nginx/modsecurity/modsecurity.conf

# Concurrent logging implies the log is stored in several files
echo "SecAuditLogStorageDir /var/log/audit/" >> /etc/nginx/modsecurity/modsecurity.conf

# Download owasp modsecurity crs
cd /etc/nginx/

git clone -b $OWASP_MODSECURITY_CRS_VERSION https://github.com/coreruleset/coreruleset
mv coreruleset owasp-modsecurity-crs
cd owasp-modsecurity-crs

mv crs-setup.conf.example crs-setup.conf
mv rules/REQUEST-900-EXCLUSION-RULES-BEFORE-CRS.conf.example rules/REQUEST-900-EXCLUSION-RULES-BEFORE-CRS.conf
mv rules/RESPONSE-999-EXCLUSION-RULES-AFTER-CRS.conf.example rules/RESPONSE-999-EXCLUSION-RULES-AFTER-CRS.conf
cd ..

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
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-922-MULTIPART-ATTACK.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-930-APPLICATION-ATTACK-LFI.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-931-APPLICATION-ATTACK-RFI.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-932-APPLICATION-ATTACK-RCE.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-933-APPLICATION-ATTACK-PHP.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-934-APPLICATION-ATTACK-NODEJS.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-942-APPLICATION-ATTACK-SQLI.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-943-APPLICATION-ATTACK-SESSION-FIXATION.conf
Include /etc/nginx/owasp-modsecurity-crs/rules/REQUEST-944-APPLICATION-ATTACK-JAVA.conf
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

# build nginx
cd "$BUILD_PATH/nginx-$NGINX_VERSION"

# apply nginx patches
for PATCH in `ls /patches`;do
  echo "Patch: $PATCH"
  if [[ "$PATCH" == *.txt ]]; then
    patch -p0 < /patches/$PATCH
  else
    patch -p1 < /patches/$PATCH
  fi
done

WITH_FLAGS="--with-debug \
  --with-compat \
  --with-pcre-jit \
  --with-http_ssl_module \
  --with-http_stub_status_module \
  --with-http_realip_module \
  --with-http_auth_request_module \
  --with-http_addition_module \
  --with-http_geoip_module \
  --with-http_gzip_static_module \
  --with-http_sub_module \
  --with-http_v2_module \
  --with-stream \
  --with-stream_ssl_module \
  --with-stream_realip_module \
  --with-stream_ssl_preread_module \
  --with-threads \
  --with-http_secure_link_module \
  --with-http_gunzip_module"

# "Combining -flto with -g is currently experimental and expected to produce unexpected results."
# https://gcc.gnu.org/onlinedocs/gcc/Optimize-Options.html
CC_OPT="-g -O2 -fPIE -fstack-protector-strong \
  -Wformat \
  -Werror=format-security \
  -Wno-deprecated-declarations \
  -fno-strict-aliasing \
  -D_FORTIFY_SOURCE=2 \
  --param=ssp-buffer-size=4 \
  -DTCP_FASTOPEN=23 \
  -fPIC \
  -I$HUNTER_INSTALL_DIR/include \
  -Wno-cast-function-type"

LD_OPT="-fPIE -fPIC -pie -Wl,-z,relro -Wl,-z,now -L$HUNTER_INSTALL_DIR/lib"

if [[ ${ARCH} != "aarch64" ]]; then
  WITH_FLAGS+=" --with-file-aio"
fi

if [[ ${ARCH} == "x86_64" ]]; then
  CC_OPT+=' -m64 -mtune=generic'
fi

WITH_MODULES=" \
  --add-module=$BUILD_PATH/ngx_devel_kit-$NDK_VERSION \
  --add-module=$BUILD_PATH/set-misc-nginx-module-$SETMISC_VERSION \
  --add-module=$BUILD_PATH/headers-more-nginx-module-$MORE_HEADERS_VERSION \
  --add-module=$BUILD_PATH/ngx_http_substitutions_filter_module-$NGINX_SUBSTITUTIONS \
  --add-module=$BUILD_PATH/lua-nginx-module-$LUA_NGX_VERSION \
  --add-module=$BUILD_PATH/stream-lua-nginx-module-$LUA_STREAM_NGX_VERSION \
  --add-module=$BUILD_PATH/lua-upstream-nginx-module-$LUA_UPSTREAM_VERSION \
  --add-module=$BUILD_PATH/nginx_ajp_module-${NGINX_AJP_VERSION} \
  --add-dynamic-module=$BUILD_PATH/nginx-http-auth-digest-$NGINX_DIGEST_AUTH \
  --add-dynamic-module=$BUILD_PATH/nginx-influxdb-module-$NGINX_INFLUXDB_VERSION \
  --add-dynamic-module=$BUILD_PATH/nginx-opentracing-$NGINX_OPENTRACING_VERSION/opentracing \
  --add-dynamic-module=$BUILD_PATH/ModSecurity-nginx-$MODSECURITY_VERSION \
  --add-dynamic-module=$BUILD_PATH/ngx_http_geoip2_module-${GEOIP2_VERSION} \
  --add-dynamic-module=$BUILD_PATH/ngx_brotli"

./configure \
  --prefix=/usr/local/nginx \
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
  --user=www-data \
  --group=www-data \
  ${WITH_MODULES}

make
make modules
make install

cd "$BUILD_PATH/lua-resty-core-$LUA_RESTY_CORE"
make install

cd "$BUILD_PATH/lua-resty-balancer-$LUA_RESTY_BALANCER"
make all
make install

export LUA_INCLUDE_DIR=/usr/local/include/luajit-2.1
ln -s $LUA_INCLUDE_DIR /usr/include/lua5.1

cd "$BUILD_PATH/lua-cjson-$LUA_CJSON_VERSION"
make all
make install

cd "$BUILD_PATH/lua-resty-cookie-$LUA_RESTY_COOKIE_VERSION"
make all
make install

cd "$BUILD_PATH/lua-resty-lrucache-$LUA_RESTY_CACHE"
make install

cd "$BUILD_PATH/lua-resty-dns-$LUA_RESTY_DNS"
make install

cd "$BUILD_PATH/lua-resty-lock-$LUA_RESTY_LOCK"
make install

# required for OCSP verification
cd "$BUILD_PATH/lua-resty-http-$LUA_RESTY_HTTP"
make install

cd "$BUILD_PATH/lua-resty-upload-$LUA_RESTY_UPLOAD_VERSION"
make install

cd "$BUILD_PATH/lua-resty-string-$LUA_RESTY_STRING_VERSION"
make install

cd "$BUILD_PATH/lua-resty-memcached-$LUA_RESTY_MEMCACHED_VERSION"
make install

cd "$BUILD_PATH/lua-resty-redis-$LUA_RESTY_REDIS_VERSION"
make install

cd "$BUILD_PATH/lua-resty-ipmatcher-$LUA_RESTY_IPMATCHER_VERSION"
INST_LUADIR=/usr/local/lib/lua make install

cd "$BUILD_PATH/lua-resty-global-throttle-$LUA_RESTY_GLOBAL_THROTTLE_VERSION"
make install

cd "$BUILD_PATH/mimalloc-$MIMALOC_VERSION"
mkdir -p out/release
cd out/release

cmake ../..

make
make install

# update image permissions
writeDirs=( \
  /etc/nginx \
  /usr/local/nginx \
  /opt/modsecurity/var/log \
  /opt/modsecurity/var/upload \
  /opt/modsecurity/var/audit \
  /var/log/audit \
  /var/log/nginx \
);

adduser -S -D -H -u 101 -h /usr/local/nginx -s /sbin/nologin -G www-data -g www-data www-data

for dir in "${writeDirs[@]}"; do
  mkdir -p ${dir};
  chown -R www-data.www-data ${dir};
done

rm -rf /etc/nginx/owasp-modsecurity-crs/.git
rm -rf /etc/nginx/owasp-modsecurity-crs/util/regression-tests

# remove .a files
find /usr/local -name "*.a" -print | xargs /bin/rm
