# Sysctl tuning

This example aims to demonstrate the use of an Init Container to adjust sysctl default values
using `kubectl patch`

```console
kubectl patch deployment -n ingress-nginx nginx-ingress-controller --patch="$(cat patch.json)"
```

