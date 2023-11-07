#!/bin/bash

# Copyright 2023 The Kubernetes Authors.
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

export NGINX_VERSION=1.25.3

# Check for recent changes: https://github.com/vision5/ngx_devel_kit/compare/v0.3.3...master
export NDK_VERSION=v0.3.3

# Check for recent changes: https://github.com/openresty/set-misc-nginx-module/compare/v0.33...master
export SETMISC_VERSION=796f5a3e518748eb29a93bd450324e0ad45b704e

# Check for recent changes: https://github.com/openresty/headers-more-nginx-module/compare/v0.34...master
export MORE_HEADERS_VERSION=v0.37

# Check for recent changes: https://github.com/atomx/nginx-http-auth-digest/compare/v1.0.0...atomx:master
export NGINX_DIGEST_AUTH=v1.0.0

# Check for recent changes: https://github.com/yaoweibin/ngx_http_substitutions_filter_module/compare/v0.6.4...master
export NGINX_SUBSTITUTIONS=e12e965ac1837ca709709f9a26f572a54d83430e

# Check for recent changes: https://github.com/SpiderLabs/ModSecurity-nginx/compare/v1.0.3...master
export MODSECURITY_VERSION=v1.0.3

# Check for recent changes: https://github.com/SpiderLabs/ModSecurity/compare/v3.0.8...v3/master
export MODSECURITY_LIB_VERSION=v3.0.11

# Check for recent changes: https://github.com/coreruleset/coreruleset/compare/v3.3.2...v3.3/master
export OWASP_MODSECURITY_CRS_VERSION=v3.3.5

# Check for recent changes: https://github.com/openresty/lua-nginx-module/compare/v0.10.25...master
export LUA_NGX_VERSION=v0.10.26

# Check for recent changes: https://github.com/openresty/stream-lua-nginx-module/compare/v0.0.13...master
export LUA_STREAM_NGX_VERSION=v0.0.14

# Check for recent changes: https://github.com/openresty/lua-upstream-nginx-module/compare/8aa93ead98ba2060d4efd594ae33a35d153589bf...master
export LUA_UPSTREAM_VERSION=542be0893543a4e42d89f6dd85372972f5ff2a36

# Check for recent changes: https://github.com/openresty/lua-cjson/compare/2.1.0.11...openresty:master
export LUA_CJSON_VERSION=2.1.0.13

# Check for recent changes: https://github.com/leev/ngx_http_geoip2_module/compare/3.4...master
export GEOIP2_VERSION=a607a41a8115fecfc05b5c283c81532a3d605425

# Check for recent changes: https://github.com/openresty/luajit2/compare/v2.1-20230410...v2.1-agentzh
export LUAJIT_VERSION=v2.1-20231117

# Check for recent changes: https://github.com/openresty/lua-resty-balancer/compare/v0.04...master
export LUA_RESTY_BALANCER=1cd4363c0a239afe4765ec607dcfbbb4e5900eea

# Check for recent changes: https://github.com/openresty/lua-resty-lrucache/compare/v0.13...master
export LUA_RESTY_CACHE=99e7578465b40f36f596d099b82eab404f2b42ed

# Check for recent changes: https://github.com/openresty/lua-resty-core/compare/v0.1.27...master
export LUA_RESTY_CORE=v0.1.28

# Check for recent changes: https://github.com/cloudflare/lua-resty-cookie/compare/v0.1.0...master
export LUA_RESTY_COOKIE_VERSION=f418d77082eaef48331302e84330488fdc810ef4

# Check for recent changes: https://github.com/openresty/lua-resty-dns/compare/v0.22...master
export LUA_RESTY_DNS=8bb53516e2933e61c317db740a9b7c2048847c2f

# Check for recent changes: https://github.com/ledgetech/lua-resty-http/compare/v0.16.1...master
export LUA_RESTY_HTTP=v0.17.1

# Check for recent changes: https://github.com/openresty/lua-resty-lock/compare/v0.09...master
export LUA_RESTY_LOCK=405d0bf4cbfa74d742c6ed3158d442221e6212a9

# Check for recent changes: https://github.com/openresty/lua-resty-upload/compare/v0.11...master
export LUA_RESTY_UPLOAD_VERSION=979372cce011f3176af3c9aff53fd0e992c4bfd3

# Check for recent changes: https://github.com/openresty/lua-resty-string/compare/v0.15...master
export LUA_RESTY_STRING_VERSION=6f1bc21d86daef804df3cc34d6427ef68da26844

# Check for recent changes: https://github.com/openresty/lua-resty-memcached/compare/v0.17...master
export LUA_RESTY_MEMCACHED_VERSION=2f02b68bf65fa2332cce070674a93a69a6c7239b

# Check for recent changes: https://github.com/openresty/lua-resty-redis/compare/v0.30...master
export LUA_RESTY_REDIS_VERSION=8641b9f1b6f75cca50c90cf8ca5c502ad8950aa8

# Check for recent changes: https://github.com/api7/lua-resty-ipmatcher/compare/v0.6.1...master
export LUA_RESTY_IPMATCHER_VERSION=3e93c53eb8c9884efe939ef070486a0e507cc5be

# Check for recent changes: https://github.com/ElvinEfendi/lua-resty-global-throttle/compare/v0.2.0...main
export LUA_RESTY_GLOBAL_THROTTLE_VERSION=v0.2.0

# Check for recent changes:  https://github.com/microsoft/mimalloc/compare/v1.7.6...master
export MIMALOC_VERSION=v2.1.2

export BUILD_PATH=/tmp/build

ARCH=$(uname -m)

get_src()
{
  hash="$1"
  url="$2"
  dest="${3-}"
  ARGS=""
  f=$(basename "$url")

  echo "Downloading $url"

  curl -sSL "$url" -o "$f"
  # echo "$hash  $f" | sha256sum -c - || exit 10
  if [ ! -z "$dest" ]; then
        mkdir ${BUILD_PATH}/${dest}
        ARGS="-C ${BUILD_PATH}/${dest} --strip-components=1"
  fi
  tar xvzf "$f" $ARGS
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

get_src aa961eafb8317e0eb8da37eb6e2c9ff42267edd18b56947384e719b85188f58b \
        "https://github.com/vision5/ngx_devel_kit/archive/$NDK_VERSION.tar.gz" "ngx_devel_kit"

get_src cd5e2cc834bcfa30149e7511f2b5a2183baf0b70dc091af717a89a64e44a2985 \
        "https://github.com/openresty/set-misc-nginx-module/archive/$SETMISC_VERSION.tar.gz" "set-misc-nginx-module"

get_src 0c0d2ced2ce895b3f45eb2b230cd90508ab2a773299f153de14a43e44c1209b3 \
        "https://github.com/openresty/headers-more-nginx-module/archive/$MORE_HEADERS_VERSION.tar.gz" "headers-more-nginx-module"

get_src f09851e6309560a8ff3e901548405066c83f1f6ff88aa7171e0763bd9514762b \
        "https://github.com/atomx/nginx-http-auth-digest/archive/$NGINX_DIGEST_AUTH.tar.gz" "nginx-http-auth-digest"

get_src a98b48947359166326d58700ccdc27256d2648218072da138ab6b47de47fbd8f \
        "https://github.com/yaoweibin/ngx_http_substitutions_filter_module/archive/$NGINX_SUBSTITUTIONS.tar.gz" "ngx_http_substitutions_filter_module"

get_src 32a42256616cc674dca24c8654397390adff15b888b77eb74e0687f023c8751b \
        "https://github.com/SpiderLabs/ModSecurity-nginx/archive/$MODSECURITY_VERSION.tar.gz" "ModSecurity-nginx"

get_src bc764db42830aeaf74755754b900253c233ad57498debe7a441cee2c6f4b07c2 \
        "https://github.com/openresty/lua-nginx-module/archive/$LUA_NGX_VERSION.tar.gz" "lua-nginx-module"

get_src 01b715754a8248cc7228e0c8f97f7488ae429d90208de0481394e35d24cef32f \
        "https://github.com/openresty/stream-lua-nginx-module/archive/$LUA_STREAM_NGX_VERSION.tar.gz" "stream-lua-nginx-module"

get_src a92c9ee6682567605ece55d4eed5d1d54446ba6fba748cff0a2482aea5713d5f \
        "https://github.com/openresty/lua-upstream-nginx-module/archive/$LUA_UPSTREAM_VERSION.tar.gz" "lua-upstream-nginx-module"

get_src 77bbcbb24c3c78f51560017288f3118d995fe71240aa379f5818ff6b166712ff \
        "https://github.com/openresty/luajit2/archive/$LUAJIT_VERSION.tar.gz" "luajit2"

get_src b6c9c09fd43eb34a71e706ad780b2ead26549a9a9f59280fe558f5b7b980b7c6 \
        "https://github.com/leev/ngx_http_geoip2_module/archive/$GEOIP2_VERSION.tar.gz" "ngx_http_geoip2_module"

get_src deb4ab1ffb9f3d962c4b4a2c4bdff692b86a209e3835ae71ebdf3b97189e40a9 \
        "https://github.com/openresty/lua-resty-upload/archive/$LUA_RESTY_UPLOAD_VERSION.tar.gz" "lua-resty-upload"

get_src bdbf271003d95aa91cab0a92f24dca129e99b33f79c13ebfcdbbcbb558129491 \
        "https://github.com/openresty/lua-resty-string/archive/$LUA_RESTY_STRING_VERSION.tar.gz" "lua-resty-string"

get_src 16d72ed133f0c6df376a327386c3ef4e9406cf51003a700737c3805770ade7c5 \
        "https://github.com/openresty/lua-resty-balancer/archive/$LUA_RESTY_BALANCER.tar.gz" "lua-resty-balancer"

get_src 39baab9e2b31cc48cecf896cea40ef6e80559054fd8a6e440cc804a858ea84d4 \
        "https://github.com/openresty/lua-resty-core/archive/$LUA_RESTY_CORE.tar.gz" "lua-resty-core"

get_src a77b9de160d81712f2f442e1de8b78a5a7ef0d08f13430ff619f79235db974d4 \
        "https://github.com/openresty/lua-cjson/archive/$LUA_CJSON_VERSION.tar.gz" "lua-cjson"

get_src 5ed48c36231e2622b001308622d46a0077525ac2f751e8cc0c9905914254baa4 \
        "https://github.com/cloudflare/lua-resty-cookie/archive/$LUA_RESTY_COOKIE_VERSION.tar.gz" "lua-resty-cookie"

get_src 573184006b98ccee2594b0d134fa4d05e5d2afd5141cbad315051ccf7e9b6403 \
        "https://github.com/openresty/lua-resty-lrucache/archive/$LUA_RESTY_CACHE.tar.gz" "lua-resty-lrucache"

get_src b4ddcd47db347e9adf5c1e1491a6279a6ae2a3aff3155ef77ea0a65c998a69c1 \
        "https://github.com/openresty/lua-resty-lock/archive/$LUA_RESTY_LOCK.tar.gz" "lua-resty-lock"

get_src 70e9a01eb32ccade0d5116a25bcffde0445b94ad35035ce06b94ccd260ad1bf0 \
        "https://github.com/openresty/lua-resty-dns/archive/$LUA_RESTY_DNS.tar.gz" "lua-resty-dns"

get_src 9fcb6db95bc37b6fce77d3b3dc740d593f9d90dce0369b405eb04844d56ac43f \
        "https://github.com/ledgetech/lua-resty-http/archive/$LUA_RESTY_HTTP.tar.gz" "lua-resty-http"

get_src 02733575c4aed15f6cab662378e4b071c0a4a4d07940c4ef19a7319e9be943d4 \
        "https://github.com/openresty/lua-resty-memcached/archive/$LUA_RESTY_MEMCACHED_VERSION.tar.gz" "lua-resty-memcached"

get_src c15aed1a01c88a3a6387d9af67a957dff670357f5fdb4ee182beb44635eef3f1 \
        "https://github.com/openresty/lua-resty-redis/archive/$LUA_RESTY_REDIS_VERSION.tar.gz" "lua-resty-redis"

get_src efb767487ea3f6031577b9b224467ddbda2ad51a41c5867a47582d4ad85d609e \
        "https://github.com/api7/lua-resty-ipmatcher/archive/$LUA_RESTY_IPMATCHER_VERSION.tar.gz" "lua-resty-ipmatcher"

get_src 0fb790e394510e73fdba1492e576aaec0b8ee9ef08e3e821ce253a07719cf7ea \
        "https://github.com/ElvinEfendi/lua-resty-global-throttle/archive/$LUA_RESTY_GLOBAL_THROTTLE_VERSION.tar.gz" "lua-resty-global-throttle"

get_src d74f86ada2329016068bc5a243268f1f555edd620b6a7d6ce89295e7d6cf18da \
        "https://github.com/microsoft/mimalloc/archive/${MIMALOC_VERSION}.tar.gz" "mimalloc"

# improve compilation times
CORES=$(($(grep -c ^processor /proc/cpuinfo) - 1))

export MAKEFLAGS=-j${CORES}
export CTEST_BUILD_FLAGS=${MAKEFLAGS}

# Install luajit from openresty fork
export LUAJIT_LIB=/usr/local/lib
export LUA_LIB_DIR="$LUAJIT_LIB/lua"
export LUAJIT_INC=/usr/local/include/luajit-2.1

cd "$BUILD_PATH/luajit2"
make CCDEBUG=-g
make install

ln -s /usr/local/bin/luajit /usr/local/bin/lua
ln -s "$LUAJIT_INC" /usr/local/include/lua

cd "$BUILD_PATH"

# Git tuning
git config --global --add core.compression -1

# Get Brotli source and deps
cd "$BUILD_PATH"
git clone --depth=100 https://github.com/google/ngx_brotli.git
cd ngx_brotli
# https://github.com/google/ngx_brotli/issues/156
git reset --hard 63ca02abdcf79c9e788d2eedcc388d2335902e52
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
  -Wno-cast-function-type"

LD_OPT="-fPIE -fPIC -pie -Wl,-z,relro -Wl,-z,now"

if [[ ${ARCH} != "aarch64" ]]; then
  WITH_FLAGS+=" --with-file-aio"
fi

if [[ ${ARCH} == "x86_64" ]]; then
  CC_OPT+=' -m64 -mtune=generic'
fi

WITH_MODULES=" \
  --add-module=$BUILD_PATH/ngx_devel_kit \
  --add-module=$BUILD_PATH/set-misc-nginx-module \
  --add-module=$BUILD_PATH/headers-more-nginx-module \
  --add-module=$BUILD_PATH/ngx_http_substitutions_filter_module \
  --add-module=$BUILD_PATH/lua-nginx-module \
  --add-module=$BUILD_PATH/stream-lua-nginx-module \
  --add-module=$BUILD_PATH/lua-upstream-nginx-module \
  --add-dynamic-module=$BUILD_PATH/nginx-http-auth-digest \
  --add-dynamic-module=$BUILD_PATH/ModSecurity-nginx \
  --add-dynamic-module=$BUILD_PATH/ngx_http_geoip2_module \
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

cd "$BUILD_PATH/lua-resty-core"
make install

cd "$BUILD_PATH/lua-resty-balancer"
make all
make install

export LUA_INCLUDE_DIR=/usr/local/include/luajit-2.1
ln -s $LUA_INCLUDE_DIR /usr/include/lua5.1

cd "$BUILD_PATH/lua-cjson"
make all
make install

cd "$BUILD_PATH/lua-resty-cookie"
make all
make install

cd "$BUILD_PATH/lua-resty-lrucache"
make install

cd "$BUILD_PATH/lua-resty-dns"
make install

cd "$BUILD_PATH/lua-resty-lock"
make install

# required for OCSP verification
cd "$BUILD_PATH/lua-resty-http"
make install

cd "$BUILD_PATH/lua-resty-upload"
make install

cd "$BUILD_PATH/lua-resty-string"
make install

cd "$BUILD_PATH/lua-resty-memcached"
make install

cd "$BUILD_PATH/lua-resty-redis"
make install

cd "$BUILD_PATH/lua-resty-ipmatcher"
INST_LUADIR=/usr/local/lib/lua make install

cd "$BUILD_PATH/lua-resty-global-throttle"
make install

cd "$BUILD_PATH/mimalloc"
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
