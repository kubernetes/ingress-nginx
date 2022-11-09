#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

declare -a mandatory
mandatory=(
  LINES
  RELEASE
)

gh pr list -R kubernetes/ingress-nginx -s merged -L ${LINES} -B main | cut -f1,2 | awk '{ printf "* [%s](https://github.com/kubernetes/ingress-nginx/pull/%s) %s\n",$1,$1, substr($0,6)}'


