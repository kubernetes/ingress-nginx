#!/bin/bash

# Copyright 2021 The Kubernetes Authors.
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

export NGINX_VERSION=1.19.10

# Check for recent changes: https://github.com/open-telemetry/opentelemetry-cpp/compare/v1.2.0...main
export OPENTELEMETRY_CPP_VERSION=1.2.0

# Check for recent changes: https://github.com/open-telemetry/opentelemetry-cpp-contrib/compare/2656a4...main
export OPENTELEMETRY_CONTRIB_COMMIT=2656a4072e257b6794da86ddd1b773b49f5517b3

export BUILD_PATH=/tmp/build

rm -rf \
   /var/cache/debconf/* \
   /var/lib/apt/lists/* \
   /var/log/* \
   /tmp/* \
   /var/tmp/*


mkdir -p /etc/nginx
mkdir --verbose -p "$BUILD_PATH"
cd "$BUILD_PATH"

apk add \
  curl \
  git \
  build-base

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


get_src e8d0290ff561986ad7cd6c33307e12e11b137186c4403a6a5ccdb4914c082d88 \
        "https://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"

get_src 360cdcbd1a235ec62119cc53956b2d31b6ff5f41d44415be53acc544709d58b8 \
        "https://github.com/open-telemetry/opentelemetry-cpp-contrib/archive/$OPENTELEMETRY_CONTRIB_COMMIT.tar.gz"

# improve compilation times
CORES=$(($(grep -c ^processor /proc/cpuinfo) - 1))

export MAKEFLAGS=-j${CORES}

apk add \
  protobuf-dev \
  grpc \
  grpc-dev \
  gtest-dev \
  c-ares-dev \
  pcre-dev

cd $BUILD_PATH
git clone --recursive https://github.com/open-telemetry/opentelemetry-cpp opentelemetry-cpp-$OPENTELEMETRY_CPP_VERSION
cd "opentelemetry-cpp-$OPENTELEMETRY_CPP_VERSION"
git checkout v$OPENTELEMETRY_CPP_VERSION
mkdir .build
cd .build

cmake -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_TESTING=OFF \
      -DWITH_EXAMPLES=OFF \
      -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
      -DWITH_OTLP=ON \
      -DWITH_OTLP_HTTP=OFF \
      ..
make
make install

# build nginx
cd "$BUILD_PATH/nginx-$NGINX_VERSION"
./configure \
  --prefix=/usr/local/nginx \
  --with-compat \
  --add-dynamic-module=$BUILD_PATH/opentelemetry-cpp-contrib-$OPENTELEMETRY_CONTRIB_COMMIT/instrumentation/nginx

make modules
mkdir -p /etc/nginx/modules
cp objs/otel_ngx_module.so /etc/nginx/modules/otel_ngx_module.so

# remove .a files
find /usr/local -name "*.a" -print | xargs /bin/rm
