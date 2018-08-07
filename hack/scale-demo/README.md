# Prometheus memory usage

This directory contains a set of test to measure the memory used by the ingress controller and the prometheus metrics.

**Requirements:**

- Running Kubernetes cluster
- kubectl
- Vegeta https://github.com/tsenart/vegeta

**Scenarios:**

The scripts in each scenario does the same thins:

- _up.sh:_

  - creates one service using the _echoheaders_ deployment scaled to 10 replicas
  - creates 500 Ingress rules and 500 services pointing to the same deployment

- _run.sh:_
  - for each ingress executes vegeta using ten connections for five seconds (~250 requests)

The only difference is `same-namespace` uses one namespace for the deployment, service and ingresses while `different-namespace` creates five hundred namespaces containing only one deployment, service and ingress.
