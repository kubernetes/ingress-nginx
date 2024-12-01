#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

if [ -n "$DEBUG" ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

export DOCKER_CLI_EXPERIMENTAL=enabled

if ! docker buildx 2>&1 >/dev/null; then
  echo "buildx not available. Docker 19.03 or higher is required with experimental features enabled"
  exit 1
fi

uname -a
docker version
docker buildx version


export BINFMT_VER="sha256:4ea31e9b91c2e7954f5f6ce991fd199a640f194ec5d7961917b5544f66062965"

# Ensure qemu is in binfmt_misc
# Docker desktop already has these in versions recent enough to have buildx
# We only need to do this setup on linux hosts
if [ "$(uname)" == 'Linux' ]; then
  # NOTE: this is pinned to a digest for a reason!
  # Note2 (@rikatz) - Removing the pin, as apparently it's breaking new alpine builds
  # docker run --rm --privileged multiarch/qemu-user-static@sha256:28ebe2e48220ae8fd5d04bb2c847293b24d7fbfad84f0b970246e0a4efd48ad6 --reset -p yes
  docker run --privileged --rm tonistiigi/binfmt@${BINFMT_VER}  --uninstall qemu-*
  docker run --rm --privileged tonistiigi/binfmt@${BINFMT_VER} --install all
fi

# Ensure we use a builder that can leverage it (the default on linux will not)
docker buildx rm ingress-nginx || true
docker buildx create --driver-opt image=moby/buildkit:v0.18.0 --driver docker-container --platform linux/amd64,linux/arm,linux/arm64 --bootstrap --use --name=ingress-nginx
