#!/bin/sh

docker run -i -v `pwd`:/gopath/src/github.com/ncabatoff/process-exporter alpine:edge /bin/sh << 'EOF'
set -ex

# Install prerequisites for the build process.
apk update
apk add git go libc-dev make

# Build the process-exporter.
cd /gopath/src/github.com/ncabatoff/process-exporter
export GOPATH=/gopath
make build
strip process-exporter
EOF
