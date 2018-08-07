#!/bin/bash

set -e

COUNT=0

while [ $COUNT -le 100 ]
do
  NAMESPACE=scale-demo-$COUNT

  kubectl delete namespace $NAMESPACE

  COUNT=$(( $COUNT + 1 ))
done
