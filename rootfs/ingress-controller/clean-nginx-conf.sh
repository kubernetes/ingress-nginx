#!/bin/bash

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

# This script removes consecutive empty lines in nginx.conf
# Using sed is more simple than using a go regex

# Sed commands:
# 1. remove the return carrier character/s
# 2. remove empty lines
# 3. replace multiple empty lines

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

sed -e 's/\r//g' | sed -e 's/^  *$/\'$'\n/g' | sed -e '/^$/{N;/^\n$/D;}' | ${SCRIPT_ROOT}/indent.sh
