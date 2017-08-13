# Dummy controller

This example contains the source code of a simple dummy controller. If you want
more details on the interface, or what the generic controller is actually doing,
please read [this doc](/docs/dev/getting-started.md). You can deploy the controller as
follows:

```console
$ kubectl create -f deployment.yaml
service "default-backend" created
deployment "dummy-ingress-controller" created

$ kubectl get po
NAME                                        READY     STATUS    RESTARTS   AGE
dummy-ingress-controller-3685541482-082nl   1/1       Running   0          10m

$ kubectl logs dummy-ingress-controller-3685541482-082nl
I0131 02:29:02.462123       1 launch.go:92] &{dummy 0.0.0 git-00000000 git://foo.bar.com}
I0131 02:29:02.462513       1 launch.go:221] Creating API server client for https://10.0.0.1:443
I0131 02:29:02.494571       1 launch.go:111] validated default/default-backend as the default backend
I0131 02:29:02.503180       1 controller.go:1038] starting Ingress controller
I0131 02:29:02.513528       1 leaderelection.go:247] lock is held by dummy-ingress-controller-3685541482-50jh0 and has not yet expired
W0131 02:29:03.510699       1 queue.go:87] requeuing kube-system/kube-scheduler, err deferring sync till endpoints controller has synced
W0131 02:29:03.514445       1 queue.go:87] requeuing kube-system/node-controller-token-826dl, err deferring sync till endpoints controller has synced
2017/01/31 02:29:12 Received OnUpdate notification
2017/01/31 02:29:12 upstream-default-backend: 10.180.1.20
```


