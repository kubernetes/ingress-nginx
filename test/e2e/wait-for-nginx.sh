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

set -e
if ! [ -z $DEBUG ]; then
	set -x
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export NAMESPACE=$1
export NAMESPACE_OVERLAY=$2

echo "deploying NGINX Ingress controller in namespace $NAMESPACE"

function on_exit {
    local error_code="$?"

    test $error_code == 0 && return;

    echo "Obtaining ingress controller pod logs..."
    kubectl logs -l app.kubernetes.io/name=ingress-nginx -n $NAMESPACE
}
trap on_exit EXIT

cat << EOF | kubectl apply --namespace=$NAMESPACE -f -
# Required for e2e tcp tests
kind: ConfigMap
apiVersion: v1
metadata:
  name: tcp-services
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx

---

# Source: nginx-ingress/templates/controller-role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
  name: nginx-ingress-controller
rules:
  - apiGroups:
      - ""
    resources:
      - namespaces
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - configmaps
      - pods
      - secrets
      - endpoints
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - services
    verbs:
      - get
      - list
      - update
      - watch
  - apiGroups:
      - extensions
      - "networking.k8s.io" # k8s 1.14+
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
      - "networking.k8s.io" # k8s 1.14+
    resources:
      - ingresses/status
    verbs:
      - update
  - apiGroups:
      - ""
    resources:
      - configmaps
    resourceNames:
      - ingress-controller-leader-nginx
    verbs:
      - get
      - update
  - apiGroups:
      - ""
    resources:
      - configmaps
    verbs:
      - create
  - apiGroups:
      - ""
    resources:
      - endpoints
    verbs:
      - create
      - get
      - update
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - create
      - patch
---
# Source: nginx-ingress/templates/controller-rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
  name: nginx-ingress-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: nginx-ingress-controller
subjects:
  - kind: ServiceAccount
    name: nginx-ingress
    namespace: $NAMESPACE

EOF

# Use the namespace overlay if it was requested
if [[ ! -z "$NAMESPACE_OVERLAY" && -d "$DIR/namespace-overlays/$NAMESPACE_OVERLAY" ]]; then
    echo "Namespace overlay $NAMESPACE_OVERLAY is being used for namespace $NAMESPACE"
    helm install nginx-ingress stable/nginx-ingress \
        --namespace=$NAMESPACE \
        --wait \
        --values "$DIR/namespace-overlays/$NAMESPACE_OVERLAY/values.yaml"
else
    cat << EOF | helm install nginx-ingress stable/nginx-ingress --namespace=$NAMESPACE --wait --values -
controller:
  image:
    repository: ingress-controller/nginx-ingress-controller
    tag: 1.0.0-dev
  scope:
    enabled: true
  config:
    worker-processes: "1"
  readinessProbe:
    initialDelaySeconds: 1
  livenessProbe:
    initialDelaySeconds: 1
  podLabels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
  service:
    type: NodePort
    labels:
      app.kubernetes.io/name: ingress-nginx
      app.kubernetes.io/part-of: ingress-nginx
  extraArgs:
    tcp-services-configmap: $NAMESPACE/tcp-services
    # e2e tests do not require information about ingress status
    update-status: "false"
  terminationGracePeriodSeconds: 1

defaultBackend:
  enabled: false

rbac:
  create: false
EOF

fi
