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

export DEBIAN_FRONTEND=noninteractive
export AR_FLAGS=cr

apt update

apt dist-upgrade --yes

add-apt-repository universe   --yes
add-apt-repository multiverse --yes

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

curl -sL -o /usr/local/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x /usr/local/bin/gimme

eval "$(gimme 1.13)"
gimme 1.13

git clone https://github.com/kubernetes/ingress-nginx

cd ingress-nginx/images/nginx

make register-qemu

PARALLELISM=${PARALLELISM:-3}

export TAG=$(git rev-parse HEAD)

# Borrowed from https://github.com/kubernetes-sigs/kind/blob/master/hack/release/build/cross.sh#L27
echo "Building in parallel for:"
# What we do here:
# - use xargs to build in parallel (-P) while collecting a combined exit code
# - use cat to supply the individual args to xargs (one line each)
# - use env -S to split the line into environment variables and execute
# - ... the build
# shellcheck disable=SC2016
if xargs -0 -n1 -P "${PARALLELISM}" bash -c 'eval $0; TAG=${TAG} make sub-container-${ARCH} > build-${ARCH}.log'; then
  echo "Docker build finished without issues" 1>&2
else
  echo "Docker build failed!" 1>&2
  cat build-amd64.log
  cat build-arm.log
  cat build-arm64.log
  exit 1
fi < <(cat <<EOF | tr '\n' '\0'
ARCH=amd64
ARCH=arm
ARCH=arm64
EOF
)

docker images

echo $QUAY_PASSWORD | sudo docker login -u $QUAY_USERNAME --password-stdin quay.io
make all-push
