# Deploying the GCE Ingress controller

This example demonstrates the deployment of a GCE Ingress controller.

Note: __all GCE/GKE clusters already have an Ingress controller running
on the master. The only reason to deploy another GCE controller is if you want
to debug or otherwise observe its operation (eg via kubectl logs). Before
deploying another one in your cluster, make sure you disable the master
controller.__

## Disabling the master controller

As of Kubernetes 1.3, GLBC runs as a static pod on the master. If you want to
totally disable it, you can ssh into the master node and delete the GLBC
manifest file found at `/etc/kubernetes/manifests/glbc.manifest`. You can also
disable it on GKE at cluster bring-up time through the `disable-addons` flag:

```console
gcloud container clusters create mycluster --network "default" --num-nodes 1 \
--machine-type n1-standard-2 --zone $ZONE \
--disable-addons HttpLoadBalancing \
--disk-size 50 --scopes storage-full
```

## Deploying a new controller

The following command deploys a GCE Ingress controller in your cluster

```console
$ kubectl create -f gce-ingress-controller.yaml
service "default-http-backend" created
replicationcontroller "l7-lb-controller" created

$ kubectl get po -l name=glbc
NAME                     READY     STATUS    RESTARTS   AGE
l7-lb-controller-1s22c   2/2       Running   0          27s
```

now you can create an Ingress and observe the controller

```console
$ kubectl create -f gce-tls-ingress.yaml
ingress "test" created

$ kubectl logs l7-lb-controller-1s22c -c l7-lb-controller
I0201 01:03:17.387548       1 main.go:179] Starting GLBC image: glbc:0.9.2, cluster name
I0201 01:03:18.459740       1 main.go:291] Using saved cluster uid "32658fa96c080068"
I0201 01:03:18.459771       1 utils.go:122] Changing cluster name from  to 32658fa96c080068
I0201 01:03:18.461652       1 gce.go:331] Using existing Token Source &oauth2.reuseTokenSource{new:google.computeSource{account:""}, mu:sync.Mutex{state:0, sema:0x0}, t:(*oauth2.Token)(nil)}
I0201 01:03:18.553142       1 cluster_manager.go:264] Created GCE client without a config file
I0201 01:03:18.553773       1 controller.go:234] Starting loadbalancer controller
I0201 01:04:58.314271       1 event.go:217] Event(api.ObjectReference{Kind:"Ingress", Namespace:"default", Name:"test", UID:"73549716-e81a-11e6-a8c5-42010af00002", APIVersion:"extensions", ResourceVersion:"673016", FieldPath:""}): type: 'Normal' reason: 'ADD' default/test
I0201 01:04:58.413616       1 instances.go:76] Creating instance group k8s-ig--32658fa96c080068 in zone us-central1-b
I0201 01:05:01.998169       1 gce.go:2084] Adding port 30301 to instance group k8s-ig--32658fa96c080068 with 0 ports
I0201 01:05:02.444014       1 backends.go:149] Creating backend for 1 instance groups, port 30301 named port &{port30301 30301 []}
I0201 01:05:02.444175       1 utils.go:495] No pod in service http-svc with node port 30301 has declared a matching readiness probe for health checks.
I0201 01:05:02.555599       1 healthchecks.go:62] Creating health check k8s-be-30301--32658fa96c080068
I0201 01:05:11.300165       1 gce.go:2084] Adding port 31938 to instance group k8s-ig--32658fa96c080068 with 1 ports
I0201 01:05:11.743914       1 backends.go:149] Creating backend for 1 instance groups, port 31938 named port &{port31938 31938 []}
I0201 01:05:11.744008       1 utils.go:495] No pod in service default-http-backend with node port 31938 has declared a matching readiness probe for health checks.
I0201 01:05:11.811972       1 healthchecks.go:62] Creating health check k8s-be-31938--32658fa96c080068
I0201 01:05:19.871791       1 loadbalancers.go:121] Creating l7 default-test--32658fa96c080068
...

$ kubectl get ing test
NAME      HOSTS     ADDRESS          PORTS     AGE
test      *         35.186.208.106   80, 443   4m

$ curl 35.186.208.106 -kL
CLIENT VALUES:
client_address=10.180.3.1
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://35.186.208.106:8080/
...
```
