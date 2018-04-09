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

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ $# -eq "1" ]
then
    export ARCH=$1
fi

source $DIR/common.sh

echo "Login to Docker Hub..."
docker login --username=$DOCKER_USERNAME --password=$DOCKER_PASSWORD >/dev/null 2>&1

if [ $# -eq "1" ]
then
    export ARCH=$1
fi

case "$COMPONENT" in
"ingress-controller")
    $DIR/ingress-controller.sh
    ;;
"nginx")
    $DIR/nginx.sh
    ;;
*)
    echo "Invalid option in environment variable COMPONENT"
    exit 1
    ;;
esac
