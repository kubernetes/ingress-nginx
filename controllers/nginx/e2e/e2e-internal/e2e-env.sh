#!/usr/bin/env bash

export MINIKUBE_VERSION=0.22.0
export K8S_VERSION=v1.7.5

export KUBECTL="/usr/local/bin/kubectl"
export MINIKUBE="/usr/local/bin/minikube"
export WAIT_FOR_DEPLOYMENT="/usr/local/bin/wait-for-deployment"

export GOOS="${GOOS:-linux}"

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
export MINIKUBE_PROFILE="ingress-e2e"

export KUBECONFIG=$HOME/.kube/config

if [ ! -e ${KUBECTL} ]; then
  echo "kubectl binary is missing. downloading binary to ${KUBECTL}"
  sudo curl -sSL http://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/${GOOS}/amd64/kubectl -o ${KUBECTL}
  sudo chmod +x ${KUBECTL}
fi

if [ ! -e ${MINIKUBE} ]; then
  echo "minikube binary is missing. downloading binary to ${MINIKUBE}"
  sudo curl -sSLo ${MINIKUBE} https://storage.googleapis.com/minikube/releases/v${MINIKUBE_VERSION}/minikube-linux-amd64
  sudo chmod +x ${MINIKUBE}
fi

if [ ! -e ${WAIT_FOR_DEPLOYMENT} ]; then
  echo "wait-for-deployment script is missing. downloading binary to ${WAIT_FOR_DEPLOYMENT}"
  # https://github.com/timoreimann/kubernetes-scripts
  sudo curl -sSLo ${WAIT_FOR_DEPLOYMENT} https://raw.githubusercontent.com/timoreimann/kubernetes-scripts/master/wait-for-deployment
  sudo chmod +x ${WAIT_FOR_DEPLOYMENT}
fi
