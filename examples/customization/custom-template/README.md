This example shows how it is possible to use a custom template

First create a configmap with a template inside running:
```
kubectl create configmap nginx-template --from-file=nginx.tmpl=../../nginx.tmpl
```

Next create the rc `kubectl create -f custom-template.yaml`
