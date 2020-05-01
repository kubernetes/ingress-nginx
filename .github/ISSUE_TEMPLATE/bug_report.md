---
name: Bug report
about: Problems and issues with code or docs
title: ''
labels: kind/bug
assignees: ''

---

<!--

Welcome to ingress-nginx!  For a smooth issue process, try to answer the following questions.
Don't worry if they're not all applicable; just try to include what you can :-)

If you need to include code snippets or logs, please put them in fenced code
blocks.  If they're super-long, please use the details tag like
<details><summary>super-long log</summary> lots of stuff </details>

-->

<!--

IMPORTANT!!!

Please complete the next sections or the issue will be closed.
This questions are the first thing we need to know to understand the context.

-->

**NGINX Ingress controller version**:

**Kubernetes version** (use `kubectl version`):

**Environment**:

- **Cloud provider or hardware configuration**:
- **OS** (e.g. from /etc/os-release):
- **Kernel** (e.g. `uname -a`):
- **Install tools**:
- **Others**:

**What happened**:

<!-- (please include exact error messages if you can) -->

**What you expected to happen**:

<!-- What do you think went wrong? -->

**How to reproduce it**:
<!---

As minimally and precisely as possible. Keep in mind we do not have access to your cluster or application.
Help up us (if possible) reproducing the issue using minikube or kind.

## Install minikube/kind

- Minikube https://minikube.sigs.k8s.io/docs/start/
- Kind https://kind.sigs.k8s.io/docs/user/quick-start/

## Install the ingress controller

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/baremetal/deploy.yaml

## Install an application that will act as default backend (is just an echo app)

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/docs/examples/http-svc.yaml

## Create an ingress (please add any additional annotation required)

echo "
  apiVersion: networking.k8s.io/v1beta1
  kind: Ingress
  metadata:
    name: foo-bar
  spec:
    rules:
    - host: foo.bar
      http:
        paths:
        - backend:
            serviceName: http-svc
            servicePort: 80
          path: /
" | kubectl apply -f -

## make a request

POD_NAME=$(k get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -o NAME)
kubectl exec -it -n ingress-nginx $POD_NAME -- curl -H 'Host: foo.bar' localhost

--->

**Anything else we need to know**:

<!-- If this is actually about documentation, add `/kind documentation` below -->

/kind bug
