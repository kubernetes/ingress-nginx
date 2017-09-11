#!/usr/bin/env bash

set -eof pipefail

. ./e2e/e2e-internal/e2e-env.sh

echo "Creating test tag for image $IMAGE:$TAG"
docker tag $IMAGE:$TAG $IMAGE:test

if [ "$TRAVIS" = true ] ; then
    echo "Using local docker graph as source for $IMAGE:test"
else
    echo "Uploading test image to minikube"
    dockerenv=$(${MINIKUBE} --profile ${MINIKUBE_PROFILE} docker-env | sed 's/export//g' | sed 's/^#.*$//g' | sed 's/"//g')
    docker save $IMAGE:test | env -i $dockerenv docker load
fi

report_error() {
    echo "unexpected error waiting for nginx ingress controller pod"
    echo "LOGS:"
    ${KUBECTL} logs -l k8s-app=nginx-ingress-controller
}

trap 'report_error $LINENO' ERR
${WAIT_FOR_DEPLOYMENT} nginx-ingress-controller
notrap

echo "Running tests..."
go test -v k8s.io/ingress/controllers/nginx/e2e/... -run ^TestIngressSuite$ --args --alsologtostderr --v=10
