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

set -o errexit
set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ "$COMPONENT" == "docs" ]; then
    echo "Skipping because we are publishing docs"
    exit 0
fi

if [ $# -eq "1" ]
then
    export ARCH=$1
fi

source $DIR/common.sh

echo "Login to quay.io..."
docker login --username=$QUAY_USERNAME --password=$QUAY_PASSWORD quay.io >/dev/null 2>&1

case "$COMPONENT" in
"ingress-controller")
    $DIR/ingress-controller.sh
    ;;
*)
    echo "Invalid option in environment variable COMPONENT"
    exit 1
    ;;
esac
