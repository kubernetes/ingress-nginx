#!/bin/bash -e

BASEDIR="$(realpath "$(dirname "$0")/..")"

echo "Cleanning dist directory..."

rm -rf ${BASEDIR}/dist