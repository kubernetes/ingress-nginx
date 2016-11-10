
To configure which services and ports will be exposed
```
kubectl create -f udp-configmap-example.yaml
```

The file `udp-configmap-example.yaml` uses a ConfigMap where the key is the external port to use and the value is
`<namespace/service name>:<service port>`
It is possible to use a number or the name of the port.

```
kubectl create -f rc-udp.yaml
```
