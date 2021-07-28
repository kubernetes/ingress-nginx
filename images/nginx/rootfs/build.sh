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

export NGINX_VERSION=1.20.1
export NDK_VERSION=0.3.1
export SETMISC_VERSION=fddc347134e97d22c34dc5189e76b0ec97bf84d7
export MORE_HEADERS_VERSION=f85af9649b858e21b400a2150a4c7b8ebd36e921
export NGINX_DIGEST_AUTH=1.0.0
export NGINX_SUBSTITUTIONS=b8a71eacc7f986ba091282ab8b1bbbc6ae1807e0
export NGINX_OPENTRACING_VERSION=2d9f40fa86380b019698f131a803f415ad24c033
export OPENTRACING_CPP_VERSION=f86b33f3d9e7322b1298ba62d5ffa7a9519c4c41
export ZIPKIN_CPP_VERSION=f69593138ff84ca2f6bc115992e18ca3d35f344a
export YAML_CPP_VERSION=b591d8ae2ad1ff373273c3e05973adf6c46abfa8
export JAEGER_VERSION=e90642a250289623fc13f650991d108e9a19b6b1
export MSGPACK_VERSION=3.3.0
export DATADOG_CPP_VERSION=5e3e3571d2341883a15fc4b5b443f930cc70e8c1
export MODSECURITY_VERSION=2497e6ac654d0b117b9534aa735b757c6b11c84f
export MODSECURITY_LIB_VERSION=v3.0.5
export OWASP_MODSECURITY_CRS_VERSION=v3.3.2
export LUA_NGX_VERSION=66a4d417d654404ade32fcc600ff1381e5e02f9c
export LUA_STREAM_NGX_VERSION=b977cee61a1cc0ea988a581cf0adbc955c9a95ab
export LUA_UPSTREAM_VERSION=8aa93ead98ba2060d4efd594ae33a35d153589bf
export LUA_CJSON_VERSION=4b350c531de3d71008c77ae94e59275b8371b4dc
export NGINX_INFLUXDB_VERSION=5b09391cb7b9a889687c0aa67964c06a2d933e8b
export GEOIP2_VERSION=a26c6beed77e81553686852dceb6c7fdacc5970d
export NGINX_AJP_VERSION=a964a0bcc6a9f2bfb82a13752d7794a36319ffac

export LUAJIT_VERSION=2.1-20210510

export LUA_RESTY_BALANCER=2e1848e729eb8558d533ce97ffdf61edead39953
export LUA_RESTY_CACHE=dacd022b4b8e3a73ed25886f2f9db043075ddb31
export LUA_RESTY_CORE=b8a4b2cfe35bcb65c2ada52c17503ece7cd3288a
export LUA_RESTY_COOKIE_VERSION=303e32e512defced053a6484bc0745cf9dc0d39e
export LUA_RESTY_DNS=0.22
export LUA_RESTY_HTTP=0.16.1
export LUA_RESTY_LOCK=ad94745a4731ee6ddbffbc3b602df25ca7c6f7ba
export LUA_RESTY_UPLOAD_VERSION=7baca92c7e741979ae5857989bbf6cc0402d6126
export LUA_RESTY_STRING_VERSION=9ace36f2dde09451c377c839117ade45eb02d460
export LUA_RESTY_MEMCACHED_VERSION=1517409a3f23116bb63f080629a6a515cf332d87
export LUA_RESTY_REDIS_VERSION=91585affcd9a8da65cb664a5b1e926dde428095a
export LUA_RESTY_IPMATCHER_VERSION=211e0d2eb8bbb558b79368f89948a0bafdc23654
export LUA_RESTY_GLOBAL_THROTTLE_VERSION=0.2.0

export BUILD_PATH=/tmp/build

ARCH=$(uname -m)

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
  curl ca-certificates \
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
get_src e462e11533d5c30baa05df7652160ff5979591d291736cfa5edb9fd2edb48c49 \
        "https://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"

get_src 0e971105e210d272a497567fa2e2c256f4e39b845a5ba80d373e26ba1abfbd85 \
        "https://github.com/simpl/ngx_devel_kit/archive/v$NDK_VERSION.tar.gz"

get_src 06218d51b0ca19ce27096499762b76746e6ab70693f279151e8964bcac2f6a6c \
        "https://github.com/openresty/set-misc-nginx-module/archive/$SETMISC_VERSION.tar.gz"

get_src 4f562f8ea1829b34366d1483e70363c7db1c9ac492c4b674832a3f36ef9dce0b \
        "https://github.com/openresty/headers-more-nginx-module/archive/$MORE_HEADERS_VERSION.tar.gz"

get_src f09851e6309560a8ff3e901548405066c83f1f6ff88aa7171e0763bd9514762b \
        "https://github.com/atomx/nginx-http-auth-digest/archive/v$NGINX_DIGEST_AUTH.tar.gz"

get_src a98b48947359166326d58700ccdc27256d2648218072da138ab6b47de47fbd8f \
        "https://github.com/yaoweibin/ngx_http_substitutions_filter_module/archive/$NGINX_SUBSTITUTIONS.tar.gz"

get_src 7cc4df20c61ad31129625cc9f177e19efbfde37b51adb3451ad81c1c1b2a6e0e \
        "https://github.com/opentracing-contrib/nginx-opentracing/archive/$NGINX_OPENTRACING_VERSION.tar.gz"

get_src cbe625cba85291712253db5bc3870d60c709acfad9a8af5a302673d3d201e3ea \
        "https://github.com/opentracing/opentracing-cpp/archive/$OPENTRACING_CPP_VERSION.tar.gz"

get_src 71de3d0658935db7ccea20e006b35e58ddc7e4c18878b9523f2addc2371e9270 \
        "https://github.com/rnburn/zipkin-cpp-opentracing/archive/$ZIPKIN_CPP_VERSION.tar.gz"

get_src fcd5e88e4f7da137d65b05a1bf680accaaa9b13443996e17da5f1c7fab02f02d \
        "https://github.com/SpiderLabs/ModSecurity-nginx/archive/$MODSECURITY_VERSION.tar.gz"

get_src 78fcfb042dfd27125715c4b9f70c73196919c03d60efb1fdf3d1b5ed0acef1ab \
        "https://github.com/jbeder/yaml-cpp/archive/$YAML_CPP_VERSION.tar.gz"

get_src 7d2cf68c16b79f99bdd097c042ac2f964d94064ebe4acb68bf6e69a5019578a4 \
        "https://github.com/jaegertracing/jaeger-client-cpp/archive/$JAEGER_VERSION.tar.gz"

get_src 754c3ace499a63e45b77ef4bcab4ee602c2c414f58403bce826b76ffc2f77d0b \
        "https://github.com/msgpack/msgpack-c/archive/cpp-$MSGPACK_VERSION.tar.gz"

get_src 59fddb95df4672fead41cd1fb068e78c56c1d7b11cec0b9ef5fc1102a9602b4b \
        "https://github.com/openresty/lua-nginx-module/archive/$LUA_NGX_VERSION.tar.gz"

get_src 473b1769fcd0cadaa79f0c9cb2ac12dc278c64aaad86e89022150f987653850f \
        "https://github.com/openresty/stream-lua-nginx-module/archive/$LUA_STREAM_NGX_VERSION.tar.gz"

get_src a92c9ee6682567605ece55d4eed5d1d54446ba6fba748cff0a2482aea5713d5f \
        "https://github.com/openresty/lua-upstream-nginx-module/archive/$LUA_UPSTREAM_VERSION.tar.gz"

get_src 1ee6dad809a5bb22efb45e6dac767f7ce544ad652d353a93d7f26b605f69fe3f \
        "https://github.com/openresty/luajit2/archive/v$LUAJIT_VERSION.tar.gz"

get_src 6294ca96b1ab25c8b8585eb0a335f632380fb6d64a12e4360793dcb7902826f4 \
        "https://github.com/DataDog/dd-opentracing-cpp/archive/$DATADOG_CPP_VERSION.tar.gz"

get_src 1af5a5632dc8b00ae103d51b7bf225de3a7f0df82f5c6a401996c080106e600e \
        "https://github.com/influxdata/nginx-influxdb-module/archive/$NGINX_INFLUXDB_VERSION.tar.gz"

get_src 4c1933434572226942c65b2f2b26c8a536ab76aa771a3c7f6c2629faa764976b \
        "https://github.com/leev/ngx_http_geoip2_module/archive/$GEOIP2_VERSION.tar.gz"

get_src 94d1512bf0e5e6ffa4eca0489db1279d51f45386fffcb8a1d2d9f7fe93518465 \
        "https://github.com/yaoweibin/nginx_ajp_module/archive/$NGINX_AJP_VERSION.tar.gz"

get_src ec8935563d70b3a45931448c64132e309b1a48f2cb17a4e326e7f0b758585b86 \
        "https://github.com/openresty/lua-resty-upload/archive/$LUA_RESTY_UPLOAD_VERSION.tar.gz"

get_src 462c6b38792bab4ca8212bdfd3f2e38f6883bb45c8fb8a03474ea813e0fab853 \
        "https://github.com/openresty/lua-resty-string/archive/$LUA_RESTY_STRING_VERSION.tar.gz"

get_src 81a2e44f66f0455fedf1273a149734bae44a67b376453c851460f1b16a3ac4f4 \
        "https://github.com/openresty/lua-resty-balancer/archive/$LUA_RESTY_BALANCER.tar.gz"

get_src 91fcc7d9b46efed4873862711a2ee626d0e87384981bb272c9ac222c631b0332 \
        "https://github.com/openresty/lua-resty-core/archive/$LUA_RESTY_CORE.tar.gz"

get_src 8d602af2669fb386931760916a39f6c9034f2363c4965f215042c086b8215238 \
        "https://github.com/openresty/lua-cjson/archive/$LUA_CJSON_VERSION.tar.gz"

get_src 5ed48c36231e2622b001308622d46a0077525ac2f751e8cc0c9905914254baa4 \
        "https://github.com/cloudflare/lua-resty-cookie/archive/$LUA_RESTY_COOKIE_VERSION.tar.gz"

get_src 23d51fd6baf111d09da8f3b3245de2b3953994d334ab8e262e0852981a8f1cb3 \
        "https://github.com/openresty/lua-resty-lrucache/archive/$LUA_RESTY_CACHE.tar.gz"

get_src 60ebd0e0773c3456f1a58eff47c427acc4069015f42be7bb70e9864733e00947 \
        "https://github.com/openresty/lua-resty-lock/archive/$LUA_RESTY_LOCK.tar.gz"

get_src 70e9a01eb32ccade0d5116a25bcffde0445b94ad35035ce06b94ccd260ad1bf0 \
        "https://github.com/openresty/lua-resty-dns/archive/v$LUA_RESTY_DNS.tar.gz"

get_src d2f867ea8cacb406172d258ba5c17ca530b5deff56d58583e6c11f4f190be015 \
        "https://github.com/ledgetech/lua-resty-http/archive/v$LUA_RESTY_HTTP.tar.gz"

get_src 3490e0cd78acf31a9ae3a688cdbd29bcc8e5c85ccccd9b14855876b7a71f324c \
        "https://github.com/openresty/lua-resty-memcached/archive/$LUA_RESTY_MEMCACHED_VERSION.tar.gz"

get_src 3069e3b9f2c57171c67fb04c9bfea0145c1556afda6d112b7d6b4f841b687ae8 \
        "https://github.com/openresty/lua-resty-redis/archive/$LUA_RESTY_REDIS_VERSION.tar.gz"

get_src b8dbd502751140993a852381bcd8e98a402454596bd91838c1e51268d42db261 \
        "https://github.com/api7/lua-resty-ipmatcher/archive/$LUA_RESTY_IPMATCHER_VERSION.tar.gz"

get_src 0fb790e394510e73fdba1492e576aaec0b8ee9ef08e3e821ce253a07719cf7ea \
        "https://github.com/ElvinEfendi/lua-resty-global-throttle/archive/v$LUA_RESTY_GLOBAL_THROTTLE_VERSION.tar.gz"

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
git clone --depth=1 -b $MODSECURITY_LIB_VERSION https://github.com/SpiderLabs/ModSecurity
cd ModSecurity/
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
  patch -p1 < /patches/$PATCH
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
CC_OPT="-g -Og -fPIE -fstack-protector-strong \
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
  CC_OPT+=' -m64 -mtune=native'
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

# mimalloc
cd "$BUILD_PATH"
git clone --depth=1 -b v1.6.7 https://github.com/microsoft/mimalloc
cd mimalloc

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

addgroup -Sg 101 www-data
adduser -S -D -H -u 101 -h /usr/local/nginx -s /sbin/nologin -G www-data -g www-data www-data

for dir in "${writeDirs[@]}"; do
  mkdir -p ${dir};
  chown -R www-data.www-data ${dir};
done

rm -rf /etc/nginx/owasp-modsecurity-crs/.git
rm -rf /etc/nginx/owasp-modsecurity-crs/util/regression-tests

# remove .a files
find /usr/local -name "*.a" -print | xargs /bin/rm
