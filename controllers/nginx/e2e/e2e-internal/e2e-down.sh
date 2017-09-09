#!/usr/bin/env bash

. ./e2e/e2e-internal/e2e-env.sh

echo "Destroying running e2e cluster..."
${MINIKUBE} --profile ${MINIKUBE_PROFILE} delete || echo "Cluster already destroyed"
