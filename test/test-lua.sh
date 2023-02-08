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

BUSTED_ARGS="-v --pattern=_test"

SHDICT_ARGS=(
    "-I" "./rootfs/etc/nginx/lua"
    "--shdict" "configuration_data 5M"
    "--shdict" "certificate_data 16M"
    "--shdict" "certificate_servers 1M"
    "--shdict" "ocsp_response_cache 1M"
    "--shdict" "balancer_ewma 1M"
    "--shdict" "quota_tracker 1M"
    "--shdict" "high_throughput_tracker 1M"
    "--shdict" "balancer_ewma_last_touched_at 1M"
    "--shdict" "balancer_ewma_locks 512k"
    "--shdict" "global_throttle_cache 5M"
    "./rootfs/etc/nginx/lua/test/run.lua"
)

if [ $# -eq 0 ]; then
    resty "${SHDICT_ARGS[@]}" ./rootfs/etc/nginx/lua/test/ ./rootfs/etc/nginx/lua/plugins/**/test
else
    resty "${SHDICT_ARGS[@]}" $@
fi
