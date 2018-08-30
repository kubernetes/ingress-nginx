# Running Systemtap in a live POD

1. Patch the ingress controller deployment to mount `/lib/module` and `/usr/src` running:

```console
kubectl patch deployment -n ingress-nginx nginx-ingress-controller --patch="$(cat deployment-patch.json)"
```

2. Run ssh to the k8s node where the ingress controller is running

3. Execute:

```console
docker exec -it --user=0 --privileged <ingress controller container> bash`
```

4. Run:

```console
cp /proc/kallsyms /boot/System.map-`uname -r`
```
