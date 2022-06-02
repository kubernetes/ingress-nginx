#!/bin/bash -e

BASEDIR="$(realpath "$(dirname "$0")/..")"

name="${1:-stratio/ingress-nginx}"

cd "$BASEDIR"

docker build --network host -t "${name}:test" -f "rootfs/Dockerfile.stratio" .
