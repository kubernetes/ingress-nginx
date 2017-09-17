#!/usr/bin/env bash

. ./e2e/e2e-internal/e2e-env.sh

mkdir -p $HOME/.kube
touch $KUBECONFIG

if [ "$TRAVIS" = true ] ; then
  sudo -E ${MINIKUBE} --profile ${MINIKUBE_PROFILE} start --vm-driver=none
else
  ${MINIKUBE} --profile ${MINIKUBE_PROFILE} start
fi

# this for loop waits until kubectl can access the api server that minikube has created
for i in {1..150} # timeout for 5 minutes
do
  $($KUBECTL get pods &> /dev/null) || true
  if [ $? -ne 1 ]; then
    break
  fi
  echo -n "."
  sleep 10
done

sleep 60

echo "Kubernetes started"
