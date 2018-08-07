#!/bin/bash

set -e

COUNT=0

while [ $COUNT -le 100 ]
do
  NAMESPACE=scale-demo-$COUNT

  echo "Installing the echoheaders application in namespace $NAMESPACE"

  kubectl create namespace $NAMESPACE || true
  kubectl apply --namespace $NAMESPACE -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/docs/examples/http-svc.yaml

  kubectl scale deployment  --namespace $NAMESPACE http-svc --replicas=2


  echo "
apiVersion: v1
kind: Service
metadata:
  labels:
    app: http-svc
  name: http-svc-$COUNT
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: http-svc
  
---

apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: batch-test-$COUNT
spec:
  rules:
  - host: test-$COUNT.foo.bar
    http:
      paths:
      -
        path: /
        backend:
          serviceName: http-svc-$COUNT
          servicePort: 80
" | kubectl create --namespace $NAMESPACE -f -

  COUNT=$(( $COUNT + 1 ))
done
