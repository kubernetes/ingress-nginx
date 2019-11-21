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

apt -q=3 update

apt -q=3 dist-upgrade --yes

add-apt-repository universe   --yes
add-apt-repository multiverse --yes

apt -q=3 update

apt -q=3 install \
  apt-transport-https \
  ca-certificates \
  curl \
  make \
  htop \
  parallel \
  software-properties-common --yes

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

add-apt-repository \
  "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) \
  stable" --yes

apt -q=3 update

apt -q=3 install docker-ce --yes

echo ${docker_password} | docker login -u ${docker_username} --password-stdin quay.io

curl -sL -o /usr/local/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x /usr/local/bin/gimme

eval "$(gimme 1.13)"

git clone https://github.com/kubernetes/ingress-nginx

cd ingress-nginx/images/nginx

make register-qemu

export TAG=$(git rev-parse HEAD)

echo "Building NGINX image in parallel:"
echo "
make sub-push-amd64
make sub-push-arm
make sub-push-arm64
" | parallel {}
