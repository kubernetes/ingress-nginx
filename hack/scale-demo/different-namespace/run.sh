#!/bin/bash

NODE_IP=10.192.0.3

NODEPORT=$(kubectl get svc -n ingress-nginx ingress-nginx -o go-template='{{range.spec.ports}}{{if .nodePort}}{{.nodePort}}{{"\n"}}{{end}}{{end}}' | head -1)

COUNT=0

while [ $COUNT -le 100 ]
do
  HOST="test-$COUNT.foo.bar"
  echo "GET http://$NODE_IP:$NODEPORT" | vegeta attack -header "Host:$HOST" -connections=10 -duration=5s | tee results.bin | vegeta report
  COUNT=$(( $COUNT + 1 ))
done
