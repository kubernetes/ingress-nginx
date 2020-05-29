#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
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

function source_tfvars() {
  eval "$(
    awk 'BEGIN {FS=OFS="="}
    !/^(#| *$)/ && /^.+=.+$/ {
      gsub(/^[ \t]+|[ \t]+$/, "", $1);
      gsub(/\./, "_", $1);
      gsub(/^[ \t]+|[ \t]+$/, "", $2);
      if ($1 && $2) print $0
    }' "$@"
  )"
}

source_tfvars /tmp/env

export DEBIAN_FRONTEND=noninteractive
export AR_FLAGS=cr

apt update
apt dist-upgrade --yes
apt update

apt install \
  apt-transport-https \
  ca-certificates \
  curl \
  make \
  htop \
  software-properties-common --yes

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

add-apt-repository \
  "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) \
  stable" --yes

apt update

apt install docker-ce --yes

export DOCKER_CLI_EXPERIMENTAL=enabled

echo ${docker_password} | docker login -u ${docker_username} --password-stdin quay.io

git clone https://github.com/kubernetes/ingress-nginx
cd ingress-nginx/images/nginx

export TAG=$(git rev-parse HEAD)

make init-docker-buildx
docker buildx use ingress-nginx --default --global

echo "Building NGINX images..."
make container
