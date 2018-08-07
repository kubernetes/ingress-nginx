#!/bin/bash

NAMESPACE=scale-demo

kubectl delete ing --namespace $NAMESPACE --all
kubectl delete svc --namespace $NAMESPACE --all
