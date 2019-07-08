# Configuration Snippets

## Ingress

The Ingress in this example adds a custom header to Nginx configuration that only applies to that specific Ingress. If you want to add headers that apply globally to all Ingresses, please have a look at [this example](../custom-headers/README.md).

```console
$ kubectl apply -f ingress.yaml
```

## Test

Check if the contents of the annotation are present in the nginx.conf file using:
`kubectl exec nginx-ingress-controller-873061567-4n3k2 -n kube-system cat /etc/nginx/nginx.conf`
