#!/usr/bin/env bash

export MINIKUBE_VERSION=0.22.0
export K8S_VERSION=v1.7.5

export PWD=`pwd`
export BASEDIR="$(dirname ${BASH_SOURCE})"
export KUBECTL="${BASEDIR}/kubectl"
export MINIKUBE="${BASEDIR}/minikube"
export GOOS="${GOOS:-linux}"

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true

export KUBECONFIG=$HOME/.kube/config

export MINIKUBE_PROFILE="ingress-e2e"

export PATH=$PATH:$BASEDIR

if [ ! -e ${KUBECTL} ]; then
  echo "kubectl binary is missing. downloading..."
  curl -sSL http://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/${GOOS}/amd64/kubectl -o ${KUBECTL}
  chmod u+x ${KUBECTL}
fi

if [ ! -e ${MINIKUBE} ]; then
  echo "minikube binary is missing. downloading..."
  curl -sSLo ${MINIKUBE} https://storage.googleapis.com/minikube/releases/v${MINIKUBE_VERSION}/minikube-linux-amd64
  chmod +x ${MINIKUBE}
fi
