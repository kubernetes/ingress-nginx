#!/usr/bin/env bash

[[ $DEBUG ]] && set -x

# include env
. hack/e2e-internal/e2e-env.sh

echo "Destroying running e2e cluster..."
${MINIKUBE} --profile ingress-e2e delete || echo "e2e cluster already destroyed"
