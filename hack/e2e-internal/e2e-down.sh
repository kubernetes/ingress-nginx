#!/usr/bin/env bash

[[ $DEBUG ]] && set -x

set -eof pipefail

# include env
. hack/e2e-internal/e2e-env.sh

echo "Destroying running docker containers..."
# do not failt if the container is not running
docker rm -f kubelet    || true
docker rm -f apiserver  || true
docker rm -f etcd       || true
