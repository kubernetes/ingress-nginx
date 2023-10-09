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

export GRPC_GIT_TAG=${GRPC_GIT_TAG:="v1.43.2"}
# Check for recent changes: https://github.com/open-telemetry/opentelemetry-cpp/compare/v1.2.0...main
export OPENTELEMETRY_CPP_VERSION=${OPENTELEMETRY_CPP_VERSION:="1.2.0"}
export INSTAL_DIR=/opt/third_party/install
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
   echo "Syntax: scriptTemplate [-h|g|o|n|p|]"
   echo "options:"
   echo "h     Print Help."
   echo "g     gRPC git tag"
   echo "o     OpenTelemetry git tag"
   echo "n     install nginx"
   echo "p     prepare"
   echo
}

prepare()
{
  apk add \
    linux-headers \
    openssl \
    curl-dev \
    openssl-dev \
    gtest-dev \
    c-ares-dev \
    pcre-dev \
    curl \
    git \
    build-base
}

install_grpc()
{
  mkdir -p $BUILD_PATH/grpc
  cd ${BUILD_PATH}/grpc
  cmake -DCMAKE_INSTALL_PREFIX=${INSTAL_DIR} \
    -DGRPC_GIT_TAG=${GRPC_GIT_TAG} /opt/third_party \
    -DgRPC_BUILD_GRPC_NODE_PLUGIN=OFF \
    -DgRPC_BUILD_GRPC_OBJECTIVE_C_PLUGIN=OFF \
    -DgRPC_BUILD_GRPC_PHP_PLUGIN=OFF \
    -DgRPC_BUILD_GRPC_PHP_PLUGIN=OFF \
    -DgRPC_BUILD_GRPC_PYTHON_PLUGIN=OFF \
    -DgRPC_BUILD_GRPC_RUBY_PLUGIN=OFF
  cmake --build . -j ${CORES} --target all install
}

install_otel()
{
  cd ${BUILD_PATH}
  export LD_LIBRARY_PATH="${LD_LIBRARY_PATH:+LD_LIBRARY_PATH:}${INSTAL_DIR}/lib:/usr/local"
  export PATH="${PATH}:${INSTAL_DIR}/bin"
  git clone --recurse-submodules -j ${CORES} --depth=1 -b \
    ${OPENTELEMETRY_CPP_VERSION} https://github.com/open-telemetry/opentelemetry-cpp.git opentelemetry-cpp-${OPENTELEMETRY_CPP_VERSION}
  cd "opentelemetry-cpp-${OPENTELEMETRY_CPP_VERSION}"
  mkdir -p .build
  cd .build

  cmake -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_POSITION_INDEPENDENT_CODE=TRUE  \
        -DWITH_ZIPKIN=OFF \
        -DWITH_JAEGER=OFF \
        -DCMAKE_INSTALL_PREFIX=${INSTAL_DIR} \
        -DBUILD_TESTING=OFF \
        -DBUILD_SHARED_LIBS=OFF \
        -DWITH_OTLP=ON \
        -DWITH_OTLP_GRPC=ON \
        -DWITH_OTLP_HTTP=OFF \
        -DWITH_ABSEIL=OFF \
        -DWITH_EXAMPLES=OFF \
        ..
  cmake --build . -j ${CORES} --target install
}

get_src()
{
  hash="$1"
  url="$2"
  f=$(basename "$url")

  echo "Downloading $url"

  curl -sSL --fail-with-body "$url" -o "$f"
  echo "$hash  $f" | sha256sum -c - || exit 10
  tar xzf "$f"
  rm -rf "$f"
}

install_nginx()
{
  export NGINX_VERSION=1.21.6

  # Check for recent changes: https://github.com/open-telemetry/opentelemetry-cpp-contrib/compare/2656a4...main
  export OPENTELEMETRY_CONTRIB_COMMIT=1ec94c82095bab61f06c7393b6f3272469d285af

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
    -DCMAKE_INSTALL_PREFIX=${INSTAL_DIR} \
    -DBUILD_SHARED_LIBS=ON \
    -DNGINX_VERSION=${NGINX_VERSION} \
    ..
  cmake --build . -j ${CORES} --target install

  mkdir -p /etc/nginx/modules
  cp ${INSTAL_DIR}/otel_ngx_module.so /etc/nginx/modules/otel_ngx_module.so

  mkdir -p ${INSTAL_DIR}/lib
}

while getopts ":hpng:o:" option; do
   case $option in
    h) # display Help
         Help
         exit;;
    g) # install gRPC with git tag
        GRPC_GIT_TAG=${OPTARG}
        install_grpc
        exit;;
    o) # install OpenTelemetry tag
        OPENTELEMETRY_CPP_VERSION=${OPTARG}
        install_otel
        exit;;
    p) # prepare
        prepare
        exit;;
    n) # install nginx
        install_nginx
        exit;;
    \?)
        Help
        exit;;
   esac
done
