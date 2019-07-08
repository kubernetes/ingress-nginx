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

if [ -z $ARCH ]; then
  echo "Environment variable ARCH is not defined. Aborting.";
  exit 0;
fi

echo "COMPONENT:                  $COMPONENT"
echo "PLATFORM:                   $ARCH"
echo "TRAVIS_REPO_SLUG:           $TRAVIS_REPO_SLUG"
echo "TRAVIS_PULL_REQUEST:        $TRAVIS_PULL_REQUEST"
echo "TRAVIS_EVENT_TYPE:          $TRAVIS_EVENT_TYPE"
echo "TRAVIS_PULL_REQUEST_BRANCH: $TRAVIS_PULL_REQUEST_BRANCH"

set -o errexit
set -o nounset
set -o pipefail

# Check if jq binary is installed
if ! [ -x "$(command -v jq)" ]; then
  echo "Installing jq..."
  sudo apt-get install -y jq
fi

if [ "$TRAVIS_REPO_SLUG" != "kubernetes/ingress-nginx" ];
then
  echo "Only builds from kubernetes/ingress-nginx repository is allowed.";
  exit 0;
fi

SKIP_MESSAGE="Publication of docker image to quay.io registry skipped."

if [ "$TRAVIS_EVENT_TYPE" != "api" ];
then
  echo "Only builds triggered from travis-ci API is allowed. $SKIP_MESSAGE";
  exit 0;
fi

if [ "$TRAVIS_PULL_REQUEST" != "false" ];
then
  echo "This is a pull request. $SKIP_MESSAGE";
  exit 0;
fi

if [ "$TRAVIS_PULL_REQUEST_BRANCH" != "" ];
then
  echo "Only images build from master branch are allowed. $SKIP_MESSAGE";
  exit 0;
fi

# variables QUAY_USERNAME and QUAY_PASSWORD are required to push docker images
if [ "$QUAY_USERNAME" == "" ];
then
  echo "Environment variable QUAY_USERNAME is missing.";
  exit 0;
fi

if [ "$QUAY_PASSWORD" == "" ];
then
  echo "Environment variable QUAY_PASSWORD is missing.";
  exit 0;
fi

function docker_tag_exists() {
    TAG=${2//\"/}
    IMAGES=$(curl -s -H "Authorization: Bearer ${QUAY_PASSWORD}" https://quay.io/api/v1/repository/$1-$3/image/ | jq '.images | sort_by(.sort_index) | .[] .tags | select(.[] !=null) | .[0]' | sed s/\"//g)
    if echo "$IMAGES" | grep -q -P "(^|\s)$TAG(?=\s|$)" ; then
        return 0
    fi

    return 1
}
