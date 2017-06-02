#!/usr/bin/env bash

go test -v k8s.io/ingress/hack/... -tags -run ^TestIngressSuite$ --args --alsologtostderr --v=10
