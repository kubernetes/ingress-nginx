#!/usr/bin/env bash

# Copyright 2015 The Kubernetes Authors All rights reserved.
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


# This test is for dev purposes.

set -e

SECRET_NAME=${SECRET_NAME:-ssl-secret}
# Name of the app in the .yaml
APP=${APP:-nginxsni}
# SNI hostnames
HOSTS=${HOSTS:-foo.bar.com}
# Should the test build and push the container via make push?
PUSH=${PUSH:-false}

# makeCerts makes certificates applying the given hostnames as CNAMEs
# $1 Name of the app that will use this secret, applied as a app= label
# $2... hostnames as described below
# Eg: makeCerts nginxsni nginx1 nginx2 nginx3
# Will generate nginx{1,2,3}.crt,.key,.json file in cwd. It's upto the caller
# to execute kubectl -f on the json file. The secret will have a label of
# app=nginxsni, so you can delete it via the cleanup function.
function makeCerts {
    local label=$1
    shift
    for h in ${@}; do
        if [ ! -f $h.json ] || [ ! -f $h.crt ] || [ ! -f $h.key ]; then
            printf "\nCreating new secrets for $h, will take ~30s\n\n"
            local cert=$h.crt key=$h.key host=$h secret=$h.json cname=$h
            if [ $h == "wildcard" ]; then
                cname=*.$h.com
            fi

            # Generate crt and key
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                    -keyout "${key}" -out "${cert}" -subj "/CN=${cname}/O=${cname}"
        fi

        cat <<EOF > secret-$SECRET_NAME-$h.json
{
    "kind": "Secret",
    "apiVersion": "v1",
    "metadata": {
        "name": "$SECRET_NAME"
    },
    "data": {
        "$h.crt": "$(cat ./$h.crt | base64)",
        "$h.key": "$(cat ./$h.key | base64)"
    }
}

EOF

    done
}

makeCerts ${APP} ${HOSTS[*]}
