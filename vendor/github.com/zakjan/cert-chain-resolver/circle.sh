#!/usr/bin/env bash

set -eu


GO_PROJECT_HOME="/home/ubuntu/.go_workspace/src/${CIRCLE_REPOSITORY_URL/https:\/\//}"

dependencies() {
    mkdir -p "${GO_PROJECT_HOME}"
    rsync -a --delete . "${GO_PROJECT_HOME}"

    cd "${GO_PROJECT_HOME}"
    go get github.com/Masterminds/glide
    glide install
}

build() {
    cd "${GO_PROJECT_HOME}"
    go build
}

test() {
    cd "${GO_PROJECT_HOME}"
    go test $(glide novendor)
    tests/run.sh
}

release() {
    cd "${GO_PROJECT_HOME}"
    mkdir out

    GOARCH="amd64"

    for GOOS in linux darwin windows; do
        echo "Building ${GOOS}_${GOARCH}"

        DIR="${CIRCLE_PROJECT_REPONAME}_${GOOS}_${GOARCH}"
        OUT="out/${DIR}/${CIRCLE_PROJECT_REPONAME}"
        if [ "${GOOS}" = "windows" ]; then
            OUT="${OUT}.exe"
        fi

        GOOS="${GOOS}" GOARCH="${GOARCH}" go build -o "${OUT}"

        cd out
        tar -czf "${DIR}.tar.gz" "${DIR}"
        rm -rf "${DIR}"
        cp "${DIR}.tar.gz" "${CIRCLE_ARTIFACTS}"
        cd ..
    done
}


case "$1" in
    dependencies)
        dependencies;;
    build)
        build;;
    test)
        test;;
    release)
        release;;
esac
