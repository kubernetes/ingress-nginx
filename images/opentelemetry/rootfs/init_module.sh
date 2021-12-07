#!/bin/sh

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

mkdir -p /modules_mount/etc
mkdir -p /modules_mount/usr/local/include

cp -R /etc/nginx/ /modules_mount/etc/nginx/
cp -R /usr/local/lib/ /modules_mount/usr/local/lib
cp -R /usr/local/include/opentelemetry/ /modules_mount/usr/local/include/opentelemetry
cp -R /usr/local/include/nlohmann/ /modules_mount/usr/local/include/nlohmann
