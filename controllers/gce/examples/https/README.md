# Simple TLS example

Create secret
```console
$ make keys secret
$ kubectl create -f /tmp/tls.json
```

Make sure you have the l7 controller running:
```console
$ kubectl --namespace=kube-system get pod -l name=glbc
NAME
l7-lb-controller-v0.6.0-1770t ...
```
Also make sure you have a [firewall rule](https://github.com/kubernetes/ingress/blob/master/controllers/gce/BETA_LIMITATIONS.md#creating-the-fir-glbc-health-checks) for the node port of the Service.

Create Ingress
```console
$ kubectl create -f tls-app.yaml
```
