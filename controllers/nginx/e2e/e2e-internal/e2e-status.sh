#!/usr/bin/env bash

set -eof pipefail

. ./e2e/e2e-internal/e2e-env.sh

${MINIKUBE} --profile ${MINIKUBE_PROFILE} status
