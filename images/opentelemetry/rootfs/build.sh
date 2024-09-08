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
set -x
# Check for recent changes: https://github.com/open-telemetry/opentelemetry-cpp/compare/v1.2.0...main
export OPENTELEMETRY_CPP_VERSION=${OPENTELEMETRY_CPP_VERSION:="v1.11.0"}
export INSTALL_DIR=/opt/third_party/install

export NGINX_VERSION=${NGINX_VERSION:="1.25.3"}
# improve compilation times
CORES=$(($(grep -c ^processor /proc/cpuinfo) - 1))

rm -rf \
   /var/cache/debconf/* \
   /var/lib/apt/lists/* \
   /var/log/* \
   /tmp/* \
   /var/tmp/*

export BUILD_PATH=/tmp/build
mkdir --verbose -p "$BUILD_PATH"

Help()
{
   # Display Help
   echo "Add description of the script functions here."
   echo
   echo "Syntax: scriptTemplate [-h|o|n|p|]"
   echo "options:"
   echo "h     Print Help."
   echo "o     OpenTelemetry git tag"
   echo "n     install nginx"
   echo "p     prepare"
   echo
}

prepare()
{
  echo "https://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories
  echo "https://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
  echo "https://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

  apk add \
    linux-headers \
    cmake \
    ninja \
    openssl \
    curl-dev \
    openssl-dev \
    gtest-dev \
    c-ares-dev \
    pcre-dev \
    curl \
    git \
    build-base \
    coreutils \
    build-base \
    openssl-dev \
    pkgconfig \
    c-ares-dev \
    re2-dev \
    grpc-dev \
    protobuf-dev \
    opentelemetry-cpp-dev

  git config --global http.version HTTP/1.1
  git config --global http.postBuffer 157286400
}

install_otel()
{
  cd ${BUILD_PATH}
  export LD_LIBRARY_PATH="${LD_LIBRARY_PATH:+LD_LIBRARY_PATH:}${INSTALL_DIR}/lib:/usr/local"
  export PATH="${PATH}:${INSTALL_DIR}/bin"
  git clone --recurse-submodules -j ${CORES} --depth=1 -b \
    ${OPENTELEMETRY_CPP_VERSION} https://github.com/open-telemetry/opentelemetry-cpp.git opentelemetry-cpp-${OPENTELEMETRY_CPP_VERSION}
  cd "opentelemetry-cpp-${OPENTELEMETRY_CPP_VERSION}"
  mkdir -p .build
  cd .build

  cmake -DCMAKE_BUILD_TYPE=Release \
        -G Ninja \
        -DCMAKE_CXX_STANDARD=17 \
        -DCMAKE_POSITION_INDEPENDENT_CODE=TRUE  \
        -DWITH_ZIPKIN=OFF \
        -DCMAKE_INSTALL_PREFIX=${INSTALL_DIR} \
        -DBUILD_TESTING=OFF \
        -DWITH_BENCHMARK=OFF \
        -DWITH_FUNC_TESTS=OFF \
        -DBUILD_SHARED_LIBS=OFF \
        -DWITH_OTLP_GRPC=ON \
        -DWITH_OTLP_HTTP=OFF \
        -DWITH_ABSEIL=ON \
        -DWITH_EXAMPLES=OFF \
        -DWITH_NO_DEPRECATED_CODE=ON \
        ..
  cmake --build . -j ${CORES} --target install
}

install_nginx()
{

  # Check for recent changes: https://github.com/open-telemetry/opentelemetry-cpp-contrib/compare/e11348bb400d5472bf1da5d6128bead66fa111ff...main
  export OPENTELEMETRY_CONTRIB_COMMIT=e11348bb400d5472bf1da5d6128bead66fa111ff

  mkdir -p /etc/nginx
  cd "$BUILD_PATH"

  # TODO fix curl
  # get_src 0528e793a97f942868616449d49326160f9cb67b2253fb2c4864603ac6ab09a9 \
  #         "https://github.com/open-telemetry/opentelemetry-cpp-contrib/archive/$OPENTELEMETRY_CONTRIB_COMMIT.tar.gz"

  git clone https://github.com/open-telemetry/opentelemetry-cpp-contrib.git \
    opentelemetry-cpp-contrib-${OPENTELEMETRY_CONTRIB_COMMIT}
  cd ${BUILD_PATH}/opentelemetry-cpp-contrib-${OPENTELEMETRY_CONTRIB_COMMIT}
  git reset --hard ${OPENTELEMETRY_CONTRIB_COMMIT}
  cd ${BUILD_PATH}/opentelemetry-cpp-contrib-${OPENTELEMETRY_CONTRIB_COMMIT}/instrumentation/nginx
  mkdir -p build
  cd build
  cmake -DCMAKE_BUILD_TYPE=Release \
        -G Ninja \
        -DCMAKE_CXX_STANDARD=17 \
        -DCMAKE_INSTALL_PREFIX=${INSTALL_DIR} \
        -DBUILD_SHARED_LIBS=ON \
        -DNGINX_VERSION=${NGINX_VERSION} \
        ..
  cmake --build . -j ${CORES} --target install

  mkdir -p /etc/nginx/modules
  cp ${INSTALL_DIR}/otel_ngx_module.so /etc/nginx/modules/otel_ngx_module.so
}

while getopts ":phn:" option; do
   case $option in
    h) # display Help
         Help
         exit;;
    p) # prepare
        prepare
        exit;;
    n) # install nginx
        NGINX_VERSION=${OPTARG}
        install_nginx
        exit;;
    \?)
        Help
        exit;;
   esac
done
