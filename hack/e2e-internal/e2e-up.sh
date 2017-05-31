#!/usr/bin/env bash

[[ $DEBUG ]] && set -x

set -eof pipefail

# include env
. hack/e2e-internal/e2e-env.sh

${MINIKUBE} --profile ingress-e2e start
${MINIKUBE} --profile ingress-e2e status

echo "Kubernetes started"
