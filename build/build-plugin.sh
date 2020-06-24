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

if [ -n "$DEBUG" ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

declare -a mandatory
mandatory=(
  PKG
  ARCH
  GIT_COMMIT
  REPO_INFO
  TAG
)

missing=false
for var in "${mandatory[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo "Environment variable $var must be set"
    missing=true
  fi
done

if [ "$missing" = true ]; then
  exit 1
fi

export CGO_ENABLED=0

release=cmd/plugin/release

function build_for_arch(){
  os=$1
  arch=$2
  extension=$3

  env GOOS="${os}" GOARCH="${arch}" go build \
    "${GOBUILD_FLAGS}" \
    -ldflags "-s -w \
      -X ${PKG}/version.RELEASE=${TAG} \
      -X ${PKG}/version.COMMIT=${GIT_COMMIT} \
      -X ${PKG}/version.REPO=${REPO_INFO}" \
    -o "${release}/kubectl-ingress_nginx${extension}" "${PKG}/cmd/plugin"

    cp LICENSE ${release}
    tar -C "${release}" -zcvf "${release}/kubectl-ingress_nginx-${os}-${arch}.tar.gz" "kubectl-ingress_nginx${extension}" LICENSE
    rm "${release}/kubectl-ingress_nginx${extension}"
    hash=$(sha256sum "${release}/kubectl-ingress_nginx-${os}-${arch}.tar.gz" | awk '{ print $1 }')
    sed -i "s/%%%shasum_${os}_${arch}%%%/${hash}/g" "${release}/ingress-nginx.yaml"
}

rm -rf "${release}"
mkdir "${release}"

cp cmd/plugin/ingress-nginx.yaml.tmpl "${release}/ingress-nginx.yaml"

sed -i "s/%%%tag%%%/${TAG}/g" ${release}/ingress-nginx.yaml

build_for_arch darwin amd64 ''
build_for_arch linux amd64 ''
build_for_arch windows amd64 '.exe'
