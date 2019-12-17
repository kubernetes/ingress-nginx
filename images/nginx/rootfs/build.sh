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

export OPENRESTY_VERSION=7508c1852265bd04fdb2dfd64989d4c490440f1a
export NGINX_DIGEST_AUTH=cd8641886c873cf543255aeda20d23e4cd603d05
export NGINX_SUBSTITUTIONS=bc58cb11844bc42735bbaef7085ea86ace46d05b
export NGINX_OPENTRACING_VERSION=0.9.0
export OPENTRACING_CPP_VERSION=1.5.1
export ZIPKIN_CPP_VERSION=0.5.2
export JAEGER_VERSION=741b1af2805388e98dbfea449f40c6d6b19c13be
export MSGPACK_VERSION=3.2.0
export DATADOG_CPP_VERSION=1.1.1
export MODSECURITY_VERSION=1.0.1
export MODSECURITY_LIB_VERSION=3.0.3
export OWASP_MODSECURITY_CRS_VERSION=3.2.0
export LUA_BRIDGE_TRACER_VERSION=0.1.1
export NGINX_INFLUXDB_VERSION=5b09391cb7b9a889687c0aa67964c06a2d933e8b
export GEOIP2_VERSION=3.2
export NGINX_AJP_VERSION=bf6cd93f2098b59260de8d494f0f4b1f11a84627
export RESTY_LUAROCKS_VERSION=3.1.3
export LUA_RESTY_BALANCER_VERSION=0.03

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

apt-get update && apt-get dist-upgrade -y

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
  lmdb-utils \
  wget \
  libcurl4-openssl-dev \
  libprotobuf-dev protobuf-compiler \
  libz-dev \
  procps \
  git g++ pkgconf flex bison doxygen libyajl-dev liblmdb-dev libtool dh-autoreconf libxml2 libpcre++-dev libxml2-dev \
  python \
  libmaxminddb-dev \
  dumb-init \
  bc \
  unzip \
  nano \
  ssdeep \
  dos2unix mercurial \
  libyaml-cpp0.6 \
  || exit 1

# https://www.mail-archive.com/debian-bugs-dist@lists.debian.org/msg1667178.html
if [[ ${ARCH} == "armv7l" ]]; then
  echo "Fixing ca-certificates"
  touch /etc/ssl/certs/ca-certificates.crt
  c_rehash
fi

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

geoip2_get "GeoLite2-City"     "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz"
geoip2_get "GeoLite2-ASN"      "http://geolite.maxmind.com/download/geoip/database/GeoLite2-ASN.tar.gz"

mkdir --verbose -p "$BUILD_PATH"
cd "$BUILD_PATH"

# download, verify and extract the source files
get_src 221cfecadd0ed2902738757e7d8fb0bc41882840f07ea87112740b44f173722f \
        "https://github.com/openresty/openresty/archive/${OPENRESTY_VERSION}.tar.gz"

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

get_src db377619a07d538bdbf328272fdec3893e6f674bdf469b3b575f778866e3ace7 \
        "https://github.com/jaegertracing/jaeger-client-cpp/archive/$JAEGER_VERSION.tar.gz"

get_src ff865a36bad5c72b8e7ebc4b7cf5f27a820fce4faff9c571c1791e3728355a39 \
        "https://github.com/msgpack/msgpack-c/archive/cpp-$MSGPACK_VERSION.tar.gz"

get_src 052fd37cd698e24ab73ee18fc3fa55acd1d43153c12a0e65b0fba0447de1117e \
        "https://github.com/DataDog/dd-opentracing-cpp/archive/v$DATADOG_CPP_VERSION.tar.gz"

get_src 6faab57557bd9cc9fc38208f6bc304c1c13cf048640779f98812cf1f9567e202 \
        "https://github.com/opentracing/lua-bridge-tracer/archive/v$LUA_BRIDGE_TRACER_VERSION.tar.gz"

get_src 1af5a5632dc8b00ae103d51b7bf225de3a7f0df82f5c6a401996c080106e600e \
        "https://github.com/influxdata/nginx-influxdb-module/archive/$NGINX_INFLUXDB_VERSION.tar.gz"

get_src 15bd1005228cf2c869a6f09e8c41a6aaa6846e4936c473106786ae8ac860fab7 \
        "https://github.com/leev/ngx_http_geoip2_module/archive/$GEOIP2_VERSION.tar.gz"

get_src 5f629a50ba22347c441421091da70fdc2ac14586619934534e5a0f8a1390a950 \
        "https://github.com/yaoweibin/nginx_ajp_module/archive/$NGINX_AJP_VERSION.tar.gz"

get_src c573435f495aac159e34eaa0a3847172a2298eb6295fcdc35d565f9f9b990513 \
        "https://luarocks.github.io/luarocks/releases/luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz"

get_src 82209d5a5d9545c6dde3db7857f84345db22162fdea9743d5e2b2094d8d407f8 \
        "https://github.com/openresty/lua-resty-balancer/archive/v${LUA_RESTY_BALANCER_VERSION}.tar.gz"

# improve compilation times
CORES=$(($(grep -c ^processor /proc/cpuinfo) - 0))

export MAKEFLAGS=-j${CORES}
export CTEST_BUILD_FLAGS=${MAKEFLAGS}
export HUNTER_JOBS_NUMBER=${CORES}
export HUNTER_KEEP_PACKAGE_SOURCES=false
export HUNTER_USE_CACHE_SERVERS=true

if [[ ${ARCH} == "armv7l" ]]; then
  export PCRE_DIR=/usr/lib/arm-linux-gnueabihf
fi

if [[ ${ARCH} == "x86_64" ]]; then
  export PCRE_DIR=/usr/lib/x86_64-linux-gnu
fi

if [[ ${ARCH} == "aarch64" ]]; then
  export PCRE_DIR=/usr/lib/aarch64-linux-gnu
fi

cd "$BUILD_PATH"

export PATH=$PATH:/usr/local/openresty/luajit

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

# build modsecurity library
cd "$BUILD_PATH"
git clone -b v$MODSECURITY_LIB_VERSION https://github.com/SpiderLabs/ModSecurity
cd ModSecurity/
git submodule init
git submodule update
sh build.sh
./configure --disable-doxygen-doc --disable-examples --disable-dependency-tracking
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
cd "$BUILD_PATH/openresty-$OPENRESTY_VERSION"

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
  --with-http_gunzip_module \
  --with-md5-asm \
  --with-sha1-asm \
  -j${CORES} "

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

WITH_MODULES="--add-module=$BUILD_PATH/nginx-http-auth-digest-$NGINX_DIGEST_AUTH \
  --add-module=$BUILD_PATH/ngx_http_substitutions_filter_module-$NGINX_SUBSTITUTIONS \
  --add-module=$BUILD_PATH/nginx-influxdb-module-$NGINX_INFLUXDB_VERSION \
  --add-dynamic-module=$BUILD_PATH/nginx-opentracing-$NGINX_OPENTRACING_VERSION/opentracing \
  --add-dynamic-module=$BUILD_PATH/ModSecurity-nginx-$MODSECURITY_VERSION \
  --add-dynamic-module=$BUILD_PATH/ngx_http_geoip2_module-${GEOIP2_VERSION} \
  --add-module=$BUILD_PATH/nginx_ajp_module-${NGINX_AJP_VERSION} \
  --add-module=$BUILD_PATH/ngx_brotli"

make

cd openresty-1.17.4.1rc0

./configure \
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

make || exit 1
make install || exit 1

cd "$BUILD_PATH/luarocks-${RESTY_LUAROCKS_VERSION}"
./configure \
  --prefix=/usr/local/openresty/luajit \
  --with-lua=/usr/local/openresty/luajit \
  --lua-suffix=jit-2.1.0-beta3 \
  --with-lua-include=/usr/local/openresty/luajit/include/luajit-2.1

make || exit 1
make install || exit 1

export PATH=$PATH:/usr/local/openresty/luajit/bin

cd /usr/local/openresty

# build and install lua-resty-waf with dependencies
export LUA_LIB_DIR=/usr/local/openresty/lualib
export LUA_INCLUDE_DIR=/tmp/build/openresty-$OPENRESTY_VERSION/openresty-1.17.4.1rc0/build/luajit-root/usr/local/openresty/luajit/include/luajit-2.1

ln -s $LUA_INCLUDE_DIR /usr/include/lua5.1

if [[ ${ARCH} != "armv7l" ]]; then
  luarocks install lrexlib-pcre 2.7.2-1 PCRE_LIBDIR=${PCRE_DIR}
fi

luarocks install lua-resty-iputils 0.3.0-1
luarocks install lua-resty-cookie 0.1.0-1

cd "$BUILD_PATH/lua-resty-balancer-$LUA_RESTY_BALANCER_VERSION"

make
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

echo "Cleaning..."

cd /

apt-mark unmarkauto \
  bash \
  curl ca-certificates \
  libgeoip1 \
  libpcre3 \
  zlib1g \
  libaio1 \
  geoip-bin \
  libyajl2 liblmdb0 libxml2 libpcre++ \
  gzip \
  openssl \
  libyaml-cpp0.6

apt-get remove -y --purge \
  build-essential \
  libgeoip-dev \
  libpcre3-dev \
  libssl-dev \
  zlib1g-dev \
  libaio-dev \
  linux-libc-dev \
  cmake \
  wget \
  patch \
  protobuf-compiler \
  python \
  xz-utils \
  bc \
  sensible-utils \
  git g++ pkgconf flex bison doxygen libyajl-dev liblmdb-dev libgeoip-dev libtool dh-autoreconf libpcre++-dev libxml2-dev

apt-get autoremove -y

# Remove configuration files left after the package removal.
# To see such packages run: apt list | grep residual
dpkg -l | grep '^rc' | awk '{print $2}' | xargs apt-get purge --yes

rm -rf "$BUILD_PATH"
rm -Rf /usr/share/man /usr/share/doc
rm -rf /tmp/* /var/tmp/*
rm -rf /var/lib/apt/lists/*
rm -rf /var/cache/apt/archives/*
rm -rf /usr/local/modsecurity/bin
rm -rf /usr/local/modsecurity/include
rm -rf /usr/local/modsecurity/lib/libmodsecurity.a

rm -rf /root/.cache

rm -rf /etc/nginx/owasp-modsecurity-crs/.git
rm -rf /etc/nginx/owasp-modsecurity-crs/util/regression-tests

rm -rf $HOME/.hunter

rm -rf $LUA_INCLUDE_DIR /usr/include/lua5.1

# update image permissions
writeDirs=( \
  /etc/nginx \
  /usr/local/openresty/nginx \
  /opt/modsecurity/var/log \
  /opt/modsecurity/var/upload \
  /opt/modsecurity/var/audit \
  /var/log/audit \
);

for dir in "${writeDirs[@]}"; do
  mkdir -p ${dir};
  chown -R www-data.www-data ${dir};
done
