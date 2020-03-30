#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
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

DIR=$(cd $(dirname "${BASH_SOURCE}")/.. && pwd -P)

RELEASE_NAME=ingress-nginx
NAMESPACE=ingress-nginx

NAMESPACE_VAR="
apiVersion: v1
kind: Namespace
metadata:
  name: $NAMESPACE
  labels:
    app.kubernetes.io/name: $RELEASE_NAME
    app.kubernetes.io/instance: ingress-nginx
"

# Baremetal
OUTPUT_FILE="${DIR}/deploy/static/provider/baremetal/deploy.yaml"
cat << EOF | helm template $RELEASE_NAME ${DIR}/charts/ingress-nginx --namespace $NAMESPACE --values - | $DIR/hack/add-namespace.py $NAMESPACE > ${OUTPUT_FILE}
controller:
  service:
    type: NodePort

  publishService:
    enabled: false

rbac:
  create: true

EOF

echo "${NAMESPACE_VAR}
$(cat ${OUTPUT_FILE})" > ${OUTPUT_FILE}

# Cloud - generic
OUTPUT_FILE="${DIR}/deploy/static/provider/cloud/deploy.yaml"
cat << EOF | helm template $RELEASE_NAME ${DIR}/charts/ingress-nginx --namespace $NAMESPACE --namespace $NAMESPACE --values - | $DIR/hack/add-namespace.py $NAMESPACE > ${OUTPUT_FILE}
controller:
  service:
    type: LoadBalancer
    externalTrafficPolicy: Local

rbac:
  create: true

EOF

echo "${NAMESPACE_VAR}
$(cat ${OUTPUT_FILE})" > ${OUTPUT_FILE}


# AWS - NLB
OUTPUT_FILE="${DIR}/deploy/static/provider/aws/deploy.yaml"
cat << EOF | helm template $RELEASE_NAME ${DIR}/charts/ingress-nginx --namespace $NAMESPACE --values - | $DIR/hack/add-namespace.py $NAMESPACE > ${OUTPUT_FILE}
controller:
  service:
    type: LoadBalancer
    externalTrafficPolicy: Local
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-backend-protocol: "tcp"
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
      service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: "true"
      # Ensure the ELB idle timeout is less than nginx keep-alive timeout. By default,
      # NGINX keep-alive is set to 75s. If using WebSockets, the value will need to be
      # increased to '3600' to avoid any potential issues.
      service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout: "60"

rbac:
  create: true

EOF

echo "${NAMESPACE_VAR}
$(cat ${OUTPUT_FILE})" > ${OUTPUT_FILE}


OUTPUT_FILE="${DIR}/deploy/static/provider/aws/deploy-tls-termination.yaml"
cat << EOF | helm template $RELEASE_NAME ${DIR}/charts/ingress-nginx --namespace $NAMESPACE --values - | $DIR/hack/add-namespace.py $NAMESPACE > ${OUTPUT_FILE}
controller:
  service:
    type: LoadBalancer
    externalTrafficPolicy: Local

    annotations:
      service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
      service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: 'true'
      service.beta.kubernetes.io/aws-load-balancer-ssl-ports: "443"
      service.beta.kubernetes.io/aws-load-balancer-ssl-cert: "arn:aws:acm:us-west-2:XXXXXXXX:certificate/XXXXXX-XXXXXXX-XXXXXXX-XXXXXXXX"
      service.beta.kubernetes.io/aws-load-balancer-proxy-protocol: "*"
      service.beta.kubernetes.io/aws-load-balancer-type: elb
      # Ensure the ELB idle timeout is less than nginx keep-alive timeout. By default,
      # NGINX keep-alive is set to 75s. If using WebSockets, the value will need to be
      # increased to '3600' to avoid any potential issues.
      service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout: "60"

    targetPorts:
      http: http
      https: http

  config:
    # Force 80 -> 443
    force-ssl-redirect: "true"
    # use-forwarded-headers: "true"

    # Obtain IP ranges from AWS and configure the defaults
    # curl https://ip-ranges.amazonaws.com/ip-ranges.json | cat ip-ranges.json | jq -r '.prefixes[] .ip_prefix'| paste -sd "," -
    # proxy-real-ip-cidr: []

rbac:
  create: true

EOF

echo "${NAMESPACE_VAR}
$(cat ${OUTPUT_FILE})" > ${OUTPUT_FILE}
