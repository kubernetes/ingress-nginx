#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
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

cd "$BUILD_PATH"
git clone --recursive --single-branch -b v0.11.1 https://github.com/p0pr0ck5/lua-resty-waf

cd lua-resty-waf

ARCH=$(uname -m)
if [[ ${ARCH} != "x86_64" ]]; then
  # replace CFLAGS
  sed -i 's/CFLAGS = -msse2 -msse3 -msse4.1 -O3/CFLAGS = -O3/' lua-aho-corasick/Makefile
  # export PCRE lib directory
  export PCRE_LIBDIR=$(find /usr/lib -name libpcre*.so* | head -1 | xargs dirname)
  luarocks install lrexlib-pcre 2.7.2-1 PCRE_LIBDIR=${PCRE_LIBDIR}
fi

curl -o 96b0a04ce62dd01b6c6c8a8c97df7ce9916d173e.patch -sSL https://github.com/p0pr0ck5/lua-resty-waf/commit/96b0a04ce62dd01b6c6c8a8c97df7ce9916d173e.patch
patch -p1 < 96b0a04ce62dd01b6c6c8a8c97df7ce9916d173e.patch

make
make install-check

# we can not use "make install" directly here because it also calls "install-deps" which requires OPM
# to avoid that we install the libraries "install-deps" would install manually

git clone -b master --single-branch https://github.com/cloudflare/lua-resty-cookie.git "$BUILD_PATH/lua-resty-cookie"
cd "$BUILD_PATH/lua-resty-cookie"
make install

git clone -b master --single-branch https://github.com/p0pr0ck5/lua-ffi-libinjection.git "$BUILD_PATH/lua-ffi-libinjection"
cd "$BUILD_PATH/lua-ffi-libinjection"
install lib/resty/*.lua "$LUA_LIB_DIR/resty/"

git clone -b master --single-branch https://github.com/cloudflare/lua-resty-logger-socket.git "$BUILD_PATH/lua-resty-logger-socket"
cd "$BUILD_PATH/lua-resty-logger-socket"
install -d "$LUA_LIB_DIR/resty/logger"
install lib/resty/logger/*.lua "$LUA_LIB_DIR/resty/logger/"

git clone -b master --single-branch https://github.com/bungle/lua-resty-random.git "$BUILD_PATH/lua-resty-random"
cd "$BUILD_PATH/lua-resty-cookie"
make install

# and do the rest of what "make instal" does
cd "$BUILD_PATH/lua-resty-waf"
install -d "$LUA_LIB_DIR/resty/waf/storage"
install -d "$LUA_LIB_DIR/rules"
install -m 644 lib/resty/*.lua "$LUA_LIB_DIR/resty/"
install -m 644 lib/resty/waf/*.lua "$LUA_LIB_DIR/resty/waf/"
install -m 644 lib/resty/waf/storage/*.lua "$LUA_LIB_DIR/resty/waf/storage/"
install -m 644 lib/*.so $LUA_LIB_DIR
install -m 644 rules/*.json "$LUA_LIB_DIR/rules/"
