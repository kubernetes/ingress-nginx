#!/usr/bin/env bash

# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

NAME="ingress-nginx"
VERSION="${1}"
echo VERSION: ${VERSION}
echo "Building HELM3 chart for ${NAME} ${VERSION} version"

# Creating a new dir in the CI build environment
CHART_TEMP_DIR="target"
mkdir -p "${CHART_TEMP_DIR}"

cp -R charts/${NAME} "${CHART_TEMP_DIR}/${NAME}"
helm package "${CHART_TEMP_DIR}/${NAME}" --app-version=${VERSION} --version=${VERSION} --destination="${CHART_TEMP_DIR}"
