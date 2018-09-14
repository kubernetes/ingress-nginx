#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
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

set -x

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source $DIR/common.sh

IMAGE=$(make -s -C $DIR/../images/nginx image-info)

if docker_tag_exists "shopify/nginx" $(echo $IMAGE | jq .tag) "$ARCH"; then
    echo "Image already published"
    exit 0
fi

echo "building nginx-$ARCH image..."
make -C $DIR/../images/nginx sub-container-$ARCH
make -C $DIR/../images/nginx sub-push-$ARCH
