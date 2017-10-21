# Custom configuration

This example aims to demonstrate the deployment of an nginx ingress controller and
use a ConfigMap to configure a custom list of headers to be passed to the upstream
server

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/docs/examples/customization/custom-headers/configmap.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/docs/examples/customization/custom-headers/custom-headers.yaml \
    | kubectl apply -f -

```

## Test

Check the contents of the configmap is present in the nginx.conf file using:
`kubectl exec nginx-ingress-controller-873061567-4n3k2 -n kube-system cat /etc/nginx/nginx.conf`
