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
fi

curl -o 96b0a04ce62dd01b6c6c8a8c97df7ce9916d173e.patch -sSL https://github.com/p0pr0ck5/lua-resty-waf/commit/96b0a04ce62dd01b6c6c8a8c97df7ce9916d173e.patch
patch -p1 < 96b0a04ce62dd01b6c6c8a8c97df7ce9916d173e.patch

make
make install
