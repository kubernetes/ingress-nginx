#!/usr/bin/env bash

[[ $DEBUG ]] && set -x

export MINIKUBE_VERSION=0.19.1
export K8S_VERSION=v1.6.4

export PWD=`pwd`
export BASEDIR="$(dirname ${BASH_SOURCE})"
export KUBECTL="${BASEDIR}/kubectl"
export MINIKUBE="${BASEDIR}/minikube"
export GOOS="${GOOS:-linux}"

if [ ! -e ${KUBECTL} ]; then
  echo "kubectl binary is missing. downloading..."
  curl -sSL http://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/${GOOS}/amd64/kubectl -o ${KUBECTL}
  chmod u+x ${KUBECTL}
fi

if [ ! -e ${MINIKUBE} ]; then
  echo "kubectl binary is missing. downloading..."
  curl -sSLo ${MINIKUBE} https://storage.googleapis.com/minikube/releases/v${MINIKUBE_VERSION}/minikube-linux-amd64
  chmod +x ${MINIKUBE}
fi
