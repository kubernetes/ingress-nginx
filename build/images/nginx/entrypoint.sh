#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
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

catch() {
  if [ "$1" == "0" ]; then
    exit 0
  fi

  echo "Error $1 occurred on $2"

  echo "Removing temporal resources..."
  terraform destroy -auto-approve \
    -var-file /root/aws.tfvars \
    -var-file /root/env.tfvars \
    -var valid_until="${EC2_VALID_UNTIL}"
}
trap 'catch $? $LINENO' ERR

terraform init

# destroy spot instance after two hours
EC2_VALID_UNTIL=$(date -d "+2 hours" +%Y-%m-%dT%H:%M:%SZ)

terraform plan \
  -var-file /root/aws.tfvars \
  -var-file /root/env.tfvars \
  -var valid_until="${EC2_VALID_UNTIL}"

terraform apply -auto-approve \
  -var-file /root/aws.tfvars \
  -var-file /root/env.tfvars \
  -var valid_until="${EC2_VALID_UNTIL}"

terraform destroy -auto-approve \
  -var-file /root/aws.tfvars \
  -var-file /root/env.tfvars \
  -var valid_until="${EC2_VALID_UNTIL}"
