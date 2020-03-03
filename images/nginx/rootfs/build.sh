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

export DEBIAN_FRONTEND=noninteractive

export NGINX_VERSION=1.17.9
export NDK_VERSION=0.3.1rc1
export SETMISC_VERSION=0.32
export MORE_HEADERS_VERSION=0.33
export NGINX_DIGEST_AUTH=cd8641886c873cf543255aeda20d23e4cd603d05
export NGINX_SUBSTITUTIONS=bc58cb11844bc42735bbaef7085ea86ace46d05b
export NGINX_OPENTRACING_VERSION=0.9.0
export OPENTRACING_CPP_VERSION=1.5.1
export ZIPKIN_CPP_VERSION=0.5.2
export JAEGER_VERSION=0.4.2
export MSGPACK_VERSION=3.2.0
export DATADOG_CPP_VERSION=1.1.3
export MODSECURITY_VERSION=1.0.1
export MODSECURITY_LIB_VERSION=6624a18a4e7fd9881a7a9b435db3e481e8e986a5
export OWASP_MODSECURITY_CRS_VERSION=3.2.0
export LUA_NGX_VERSION=0.10.15
export LUA_STREAM_NGX_VERSION=0.0.7
export LUA_UPSTREAM_VERSION=0.07
export LUA_BRIDGE_TRACER_VERSION=0.1.1
export NGINX_INFLUXDB_VERSION=5b09391cb7b9a889687c0aa67964c06a2d933e8b
export GEOIP2_VERSION=3.3
export NGINX_AJP_VERSION=bf6cd93f2098b59260de8d494f0f4b1f11a84627
export RESTY_LUAROCKS_VERSION=3.1.3
export LUAJIT_VERSION=38cb695de87cfeadcba5eeaf57f39e41d529aa1f
export LUA_RESTY_BALANCER=0.03
export LUA_RESTY_CORE=0.1.17
export LUA_CJSON_VERSION=2.1.0.7
export LUA_RESTY_COOKIE_VERSION=766ad8c15e498850ac77f5e0265f1d3f30dc4027

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

apk update
apk upgrade

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
  geoip-dev \
  patch \
  libaio-dev \
  openssl \
  cmake \
  util-linux \
  lmdb-tools \
  wget \
  curl-dev \
  libprotobuf \
  git g++ pkgconf flex bison doxygen yajl-dev lmdb-dev libtool autoconf libxml2 pcre-dev libxml2-dev \
  python \
  libmaxminddb-dev \
  bc \
  unzip \
  dos2unix mercurial \
  yaml-cpp

mkdir -p /etc/nginx

# Get the GeoIP data
GEOIP_FOLDER=/etc/nginx/geoip
mkdir -p $GEOIP_FOLDER

function geoip2_get {
  wget -O $GEOIP_FOLDER/$1.tar.gz $2 || { echo "Could not download $1, exiting." ; exit 1; }
  mkdir $GEOIP_FOLDER/$1 \
    && tar xf $GEOIP_FOLDER/$1.tar.gz -C $GEOIP_FOLDER/$1 --strip-components 1 \
    && mv $GEOIP_FOLDER/$1/$1.mmdb $GEOIP_FOLDER/$1.mmdb \
    && rm -rf $GEOIP_FOLDER/$1 \
    && rm -rf $GEOIP_FOLDER/$1.tar.gz
}

#geoip2_get "GeoLite2-City"     "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz"
#geoip2_get "GeoLite2-ASN"      "http://geolite.maxmind.com/download/geoip/database/GeoLite2-ASN.tar.gz"

mkdir --verbose -p "$BUILD_PATH"
cd "$BUILD_PATH"

# download, verify and extract the source files
get_src 7dd65d405c753c41b7fdab9415cfb4bdbaf093ec6d9f7432072d52cb7bcbb689 \
        "https://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"

get_src 49f50d4cd62b166bc1aaf712febec5e028d9f187cedbc27a610dfd01bdde2d36 \
        "https://github.com/simpl/ngx_devel_kit/archive/v$NDK_VERSION.tar.gz"

get_src f1ad2459c4ee6a61771aa84f77871f4bfe42943a4aa4c30c62ba3f981f52c201 \
        "https://github.com/openresty/set-misc-nginx-module/archive/v$SETMISC_VERSION.tar.gz"

get_src a3dcbab117a9c103bc1ea5200fc00a7b7d2af97ff7fd525f16f8ac2632e30fbf \
        "https://github.com/openresty/headers-more-nginx-module/archive/v$MORE_HEADERS_VERSION.tar.gz"

get_src fe683831f832aae4737de1e1026a4454017c2d5f98cb88b08c5411dc380062f8 \
        "https://github.com/atomx/nginx-http-auth-digest/archive/$NGINX_DIGEST_AUTH.tar.gz"

get_src 618551948ab14cac51d6e4ad00452312c7b09938f59ebff4f93875013be31f2d \
        "https://github.com/yaoweibin/ngx_http_substitutions_filter_module/archive/$NGINX_SUBSTITUTIONS.tar.gz"

get_src 4fc410d7aef0c8a6371afa9f249d2c6cec50ea88785d05052f8f457c35b69c18 \
        "https://github.com/opentracing-contrib/nginx-opentracing/archive/v$NGINX_OPENTRACING_VERSION.tar.gz"

get_src 015c4187f7a6426a2b5196f0ccd982aa87f010cf61f507ae3ce5c90523f92301 \
        "https://github.com/opentracing/opentracing-cpp/archive/v$OPENTRACING_CPP_VERSION.tar.gz"

get_src 30affaf0f3a84193f7127cc0135da91773ce45d902414082273dae78914f73df \
        "https://github.com/rnburn/zipkin-cpp-opentracing/archive/v$ZIPKIN_CPP_VERSION.tar.gz"

get_src c969a78659bb47c84929de0b9adc1f8c512a51ec9dd3b162cb568ae228d3d59e \
        "https://github.com/SpiderLabs/ModSecurity-nginx/archive/v$MODSECURITY_VERSION.tar.gz"

get_src 21257af93a64fee42c04ca6262d292b2e4e0b7b0660c511db357b32fd42ef5d3 \
        "https://github.com/jaegertracing/jaeger-client-cpp/archive/v$JAEGER_VERSION.tar.gz"

get_src ff865a36bad5c72b8e7ebc4b7cf5f27a820fce4faff9c571c1791e3728355a39 \
        "https://github.com/msgpack/msgpack-c/archive/cpp-$MSGPACK_VERSION.tar.gz"

get_src 7d5f3439c8df56046d0564b5857fd8a30296ab1bd6df0f048aed7afb56a0a4c2 \
        "https://github.com/openresty/lua-nginx-module/archive/v$LUA_NGX_VERSION.tar.gz"

get_src 99c47c75c159795c9faf76bbb9fa58e5a50b75286c86565ffcec8514b1c74bf9 \
        "https://github.com/openresty/stream-lua-nginx-module/archive/v$LUA_STREAM_NGX_VERSION.tar.gz"

get_src 2a69815e4ae01aa8b170941a8e1a10b6f6a9aab699dee485d58f021dd933829a \
        "https://github.com/openresty/lua-upstream-nginx-module/archive/v$LUA_UPSTREAM_VERSION.tar.gz"

get_src 7df70318762f4150e6fe27dd1838b4b89a24ed9351c82d0b332d7d8457dd1b95 \
        "https://github.com/openresty/luajit2/archive/$LUAJIT_VERSION.tar.gz"

get_src 6dc1088ab7f788b6c849fbaa6300517c8fdf88991a70b778be79c284c36857bf \
        "https://github.com/DataDog/dd-opentracing-cpp/archive/v$DATADOG_CPP_VERSION.tar.gz"

get_src 6faab57557bd9cc9fc38208f6bc304c1c13cf048640779f98812cf1f9567e202 \
        "https://github.com/opentracing/lua-bridge-tracer/archive/v$LUA_BRIDGE_TRACER_VERSION.tar.gz"

get_src 1af5a5632dc8b00ae103d51b7bf225de3a7f0df82f5c6a401996c080106e600e \
        "https://github.com/influxdata/nginx-influxdb-module/archive/$NGINX_INFLUXDB_VERSION.tar.gz"

get_src 41378438c833e313a18869d0c4a72704b4835c30acaf7fd68013ab6732ff78a7 \
        "https://github.com/leev/ngx_http_geoip2_module/archive/$GEOIP2_VERSION.tar.gz"

get_src 5f629a50ba22347c441421091da70fdc2ac14586619934534e5a0f8a1390a950 \
        "https://github.com/yaoweibin/nginx_ajp_module/archive/$NGINX_AJP_VERSION.tar.gz"

get_src c573435f495aac159e34eaa0a3847172a2298eb6295fcdc35d565f9f9b990513 \
        "https://luarocks.github.io/luarocks/releases/luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz"

get_src 5d16e623d17d4f42cc64ea9cfb69ca960d313e12f5d828f785dd227cc483fcbd \
        "https://github.com/openresty/lua-resty-upload/archive/v0.10.tar.gz"

get_src 095615fe94e64615c4a27f4f4475b91c047cf8d10bc2dbde8d5ba6aa625fc5ab \
        "https://github.com/openresty/lua-resty-string/archive/v0.11.tar.gz"

get_src 82209d5a5d9545c6dde3db7857f84345db22162fdea9743d5e2b2094d8d407f8 \
        "https://github.com/openresty/lua-resty-balancer/archive/v$LUA_RESTY_BALANCER.tar.gz"

get_src 8f5f76d2689a3f6b0782f0a009c56a65e4c7a4382be86422c9b3549fe95b0dc4 \
        "https://github.com/openresty/lua-resty-core/archive/v$LUA_RESTY_CORE.tar.gz"

get_src 59d2f18ecadba48be61061004c8664eaed1111a3372cd2567cb24c5a47eb41fe \
        "https://github.com/openresty/lua-cjson/archive/$LUA_CJSON_VERSION.tar.gz"

get_src f818b5cef0881e5987606f2acda0e491531a0cb0c126d8dca02e2343edf641ef \
        "https://github.com/cloudflare/lua-resty-cookie/archive/$LUA_RESTY_COOKIE_VERSION.tar.gz"

# improve compilation times
CORES=$(($(grep -c ^processor /proc/cpuinfo) - 0))

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

cd "$BUILD_PATH"

# install openresty-gdb-utils
cd /
git clone --depth=1 https://github.com/openresty/openresty-gdb-utils.git
cat > ~/.gdbinit << EOF
directory /openresty-gdb-utils

py import sys
py sys.path.append("/openresty-gdb-utils")

source luajit20.gdb
source ngx-lua.gdb
source luajit21.py
source ngx-raw-req.py
set python print-stack full
EOF

# build opentracing lib
cd "$BUILD_PATH/opentracing-cpp-$OPENTRACING_CPP_VERSION"
mkdir .build
cd .build

cmake   -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_CXX_FLAGS="-fPIC" \
        -DBUILD_TESTING=OFF \
        -DBUILD_MOCKTRACER=OFF \
        ..

make
make install

if [[ ${ARCH} != "armv7l" ]]; then
  # build jaeger lib
  cd "$BUILD_PATH/jaeger-client-cpp-$JAEGER_VERSION"
  sed -i 's/-Werror/-Wno-psabi/' CMakeLists.txt

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
        -DJAEGERTRACING_WITH_YAML_CPP=ON ..

  make
  make install

  export HUNTER_INSTALL_DIR=$(cat _3rdParty/Hunter/install-root-dir) \

  mv libjaegertracing_plugin.so /usr/local/lib/libjaegertracing_plugin.so
fi

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
      -DBUILD_SHARED_LIBS=ON \
      -DBUILD_PLUGIN=ON \
      -DBUILD_TESTING=OFF ..

make
make install

# build msgpack lib
cd "$BUILD_PATH/msgpack-c-cpp-$MSGPACK_VERSION"

mkdir .build
cd .build
cmake -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_CXX_FLAGS="-fPIC" \
        -DBUILD_SHARED_LIBS=OFF \
        -DBUILD_TESTING=OFF \
        -DBUILD_MOCKTRACER=OFF \
        ..

make
make install

# build datadog lib
cd "$BUILD_PATH/dd-opentracing-cpp-$DATADOG_CPP_VERSION"

mkdir .build
cd .build
cmake ..

make
make install

# Get Brotli source and deps
cd "$BUILD_PATH"
git clone --depth=1 https://github.com/google/ngx_brotli.git
cd ngx_brotli
git submodule init
git submodule update

cd "$BUILD_PATH"
git clone https://github.com/ssdeep-project/ssdeep
cd ssdeep/

./bootstrap
./configure

make
make install

# build modsecurity library
cd "$BUILD_PATH"
git clone https://github.com/SpiderLabs/ModSecurity
cd ModSecurity/
git checkout $MODSECURITY_LIB_VERSION
git submodule init
git submodule update

sh build.sh

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

git clone -b v$OWASP_MODSECURITY_CRS_VERSION https://github.com/SpiderLabs/owasp-modsecurity-crs
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
  --with-http_dav_module \
  --with-http_geoip_module \
  --with-http_gzip_static_module \
  --with-http_sub_module \
  --with-http_v2_module \
  --with-stream \
  --with-stream_ssl_module \
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
  -Wno-cast-function-type"

LD_OPT="-fPIE -fPIC -pie -Wl,-z,relro -Wl,-z,now"

if [[ ${ARCH} != "armv7l" ]]; then
  CC_OPT+=" -I$HUNTER_INSTALL_DIR/include"
  LD_OPT+=" -L$HUNTER_INSTALL_DIR/lib"
fi

if [[ ${ARCH} != "aarch64" ]]; then
  WITH_FLAGS+=" --with-file-aio"
fi

if [[ ${ARCH} == "x86_64" ]]; then
  CC_OPT+=' -m64 -mtune=native'
fi

WITH_MODULES="--add-module=$BUILD_PATH/ngx_devel_kit-$NDK_VERSION \
  --add-module=$BUILD_PATH/set-misc-nginx-module-$SETMISC_VERSION \
  --add-module=$BUILD_PATH/headers-more-nginx-module-$MORE_HEADERS_VERSION \
  --add-module=$BUILD_PATH/nginx-http-auth-digest-$NGINX_DIGEST_AUTH \
  --add-module=$BUILD_PATH/ngx_http_substitutions_filter_module-$NGINX_SUBSTITUTIONS \
  --add-module=$BUILD_PATH/lua-nginx-module-$LUA_NGX_VERSION \
  --add-module=$BUILD_PATH/stream-lua-nginx-module-$LUA_STREAM_NGX_VERSION \
  --add-module=$BUILD_PATH/lua-upstream-nginx-module-$LUA_UPSTREAM_VERSION \
  --add-module=$BUILD_PATH/nginx-influxdb-module-$NGINX_INFLUXDB_VERSION \
  --add-dynamic-module=$BUILD_PATH/nginx-opentracing-$NGINX_OPENTRACING_VERSION/opentracing \
  --add-dynamic-module=$BUILD_PATH/ModSecurity-nginx-$MODSECURITY_VERSION \
  --add-dynamic-module=$BUILD_PATH/ngx_http_geoip2_module-${GEOIP2_VERSION} \
  --add-module=$BUILD_PATH/nginx_ajp_module-${NGINX_AJP_VERSION} \
  --add-module=$BUILD_PATH/ngx_brotli"

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
make install

cd "$BUILD_PATH/luarocks-${RESTY_LUAROCKS_VERSION}"
./configure \
  --lua-suffix=jit-2.1.0-beta3 \
  --with-lua-include=/usr/local/include/luajit-2.1

make
make install

export LUA_INCLUDE_DIR=/usr/local/include/luajit-2.1

ln -s $LUA_INCLUDE_DIR /usr/include/lua5.1

if [[ ${ARCH} != "armv7l" ]]; then
  luarocks install lrexlib-pcre 2.7.2-1
fi

cd "$BUILD_PATH/lua-resty-core-$LUA_RESTY_CORE"
make install

cd "$BUILD_PATH/lua-resty-balancer-$LUA_RESTY_BALANCER"
make all
make install

cd "$BUILD_PATH/lua-cjson-$LUA_CJSON_VERSION"
make all
make install

cd "$BUILD_PATH/lua-resty-cookie-$LUA_RESTY_COOKIE_VERSION"
make all
make install

luarocks install lua-resty-iputils 0.3.0-1
luarocks install lua-resty-lrucache 0.09-2
luarocks install lua-resty-lock 0.08-0
luarocks install lua-resty-dns 0.21-1

# required for OCSP verification
luarocks install lua-resty-http

cd "$BUILD_PATH/lua-resty-upload-0.10"
make install

cd "$BUILD_PATH/lua-resty-string-0.11"
make install

# build Lua bridge tracer
cd "$BUILD_PATH/lua-bridge-tracer-$LUA_BRIDGE_TRACER_VERSION"
mkdir .build
cd .build
cmake ..
make
make install

# mimalloc
cd "$BUILD_PATH"
git clone https://github.com/microsoft/mimalloc
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
rm -rf /usr/local/modsecurity/lib/libmodsecurity.a
