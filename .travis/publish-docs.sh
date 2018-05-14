#!/usr/bin/env bash

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

if ! [ -z $DEBUG ]; then
  set -x
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ "$COMPONENT" != "docs" ]; then
    echo "This task runs only to publish docs"
    exit 0
fi

make -C ${DIR}/.. build-docs

git config --global user.email "travis@travis-ci.com"
git config --global user.name "Travis Bot"

git clone --branch=gh-pages --depth=1 https://${GH_REF} ${DIR}/gh-pages
cd ${DIR}/gh-pages

git rm -r .

cp -r ${DIR}/../site/* .

git add .
git commit -m "Deploy GitHub Pages"
git push --force --quiet "https://${GH_TOKEN}@${GH_REF}" gh-pages > /dev/null 2>&1
