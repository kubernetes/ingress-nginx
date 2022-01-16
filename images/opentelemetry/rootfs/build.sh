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

export NGINX_VERSION=1.19.9

# Check for recent changes: https://github.com/open-telemetry/opentelemetry-cpp/compare/v1.0.0...main
export OPENTELEMETRY_CPP_VERSION=1.0.0

# Check for recent changes: https://github.com/open-telemetry/opentelemetry-cpp-contrib/compare/f4850...main
export OPENTELEMETRY_CONTRIB_COMMIT=f48500884b1b32efc456790bbcdc2e6cf7a8e630

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


get_src e462e11533d5c30baa05df7652160ff5979591d291736cfa5edb9fd2edb48c49 \
        "https://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"

get_src 45c52498788e47131b20a4786dbb08f4390b8cb419bd3d61c88b503cafff3324 \
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
