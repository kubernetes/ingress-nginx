#!/usr/bin/env bash

# Copyright 2015 The Kubernetes Authors.
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


# https://www.openssl.org/docs/manmaster/apps/dhparam.html
# this command generates a key used to get "Perfect Forward Secrecy" in nginx
# https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
openssl dhparam -out dhparam.pem 4096

cat <<EOF > dhparam-example.yaml
{
    "kind": "Secret",
    "apiVersion": "v1",
    "metadata": {
        "name": "dhparam-example"
    },
    "data": {
        "dhparam.pem": "$(cat ./dhparam.pem | base64)"
    }
}

EOF
