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

if ! [ -z $DEBUG ]; then
	set -x
fi

set -o errexit
set -o pipefail

if [ "$TRAVIS_CI_TOKEN" == "" ];
then
	echo "Environment variable TRAVIS_CI_TOKEN is missing.";
	exit 1;
fi

function publish() {

body=$(cat <<EOF
{
  "request": {
    "branch": "master",
		"message": "Publishing image for component $2 to quay.io",
    "config": {
      "merge_mode": "deep_merge",
			"env": {
      	"matrix" : {
        	"COMPONENT": "$2"
       	}
      }
    }
  }
}
EOF
)

	curl -s -X POST \
		-H "Content-Type: application/json" \
		-H "Accept: application/json" \
		-H "Travis-API-Version: 3" \
		-H "Authorization: token $1" \
		--data "$body" \
		https://api.travis-ci.org/repo/kubernetes%2Fingress-nginx/requests
}

case "$1" in
	ingress-controller|nginx)
		publish $TRAVIS_CI_TOKEN $1
	;;
	*)
		echo "Invalid publish option"
		exit 1
	;;
esac
