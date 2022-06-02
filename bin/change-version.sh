#!/bin/bash -e

BASEDIR="$(realpath "$(dirname "$0")/..")"

cd $BASEDIR

if [[ -z "$1" ]]; then
	VERSION=$(cat $BASEDIR/VERSION)
else
	VERSION=$1
fi

echo "Modifying ingress-nginx version to: $1"
echo $VERSION > VERSION
