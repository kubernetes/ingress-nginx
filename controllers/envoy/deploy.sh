#!/bin/bash

set -euo pipefail

COMMIT=`git rev-parse --verify HEAD`
sed "s/<<COMMIT>>/$COMMIT/" deployment.yaml > docker-build/deployment.yaml
kubectl apply -f docker-build/deployment.yaml
