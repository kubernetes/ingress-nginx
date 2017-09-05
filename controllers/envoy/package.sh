#!/bin/bash

# Build a docker container containing envoy-controller + envoy
set -euo pipefail

rm -rf docker-build
mkdir docker-build/

git commit -a -m "f WIP"

COMMIT=`git rev-parse --verify HEAD`

echo "Building envoy-controller"
go build -o docker-build/envoy-controller


echo "Packaging inside a docker"
cat <<EOF -> docker-build/dockerfile
FROM lyft/envoy-alpine:3769d9f90d2ac362d7d26176f132db996efcc4d6
COPY envoy-controller /envoy-controller
CMD /envoy-controller
EOF

# docker build docker-build -t <TODO>
# docker push <TODO>

./deploy.sh
