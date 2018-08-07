#!/bin/bash

NAMESPACE=scale-demo

echo "Installing the echoheaders application in namespace $NAMESPACE"

kubectl create namespace $NAMESPACE
kubectl apply --namespace $NAMESPACE -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/docs/examples/http-svc.yaml

kubectl scale deployment  --namespace $NAMESPACE http-svc --replicas=10

COUNT=0

while [ $COUNT -le 500 ]
do
echo "
apiVersion: v1
kind: Service
metadata:
  labels:
    app: http-svc
  name: http-svc-$COUNT
  namespace: $NAMESPACE
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
  namespace: $NAMESPACE
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
" | kubectl create -f -

COUNT=$(( $COUNT + 1 ))
done
