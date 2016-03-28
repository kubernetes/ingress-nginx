# GLBC: Beta limitations

As of the Kubernetes 1.2 release, the GCE L7 Loadbalancer controller is still a *beta* product. We expect it to go GA in 1.3.

This is a list of beta limitations:

* [Firewalls](#creating-the-firewall-rule-for-glbc-health-checks): You must create the firewall-rule required for GLBC's health checks to succeed.
* [UIDs](#running-multiple-loadbalanced-clusters-in-the-same-gce-project): If you're creating multiple clusters that will use Ingress within a single GCE project, you must assign a UID to GLBC so it doesn't stomp on resources from another cluster.
* [Health Checks](#health-checks): All Kubernetes services must serve a 200 page on '/', or whatever custom value you've specified through GLBC's `--health-check-path argument`.
* [IPs](#static-and-ephemeral-ips): Creating a simple HTTP Ingress will allocate an ephemeral IP. Creating an Ingress with a TLS section will allocate a static IP.
* [Latency](#latency): GLBC is not built for performance. Creating many Ingresses at a time can overwhelm it. It won't fall over, but will take its own time to churn through the Ingress queue.
* [Quota](#quota): By default, GCE projects are granted a quota of 3 Backend Services. This is insufficient for most Kubernetes clusters.
* [Oauth scopes](https://cloud.google.com/compute/docs/authentication): By default GKE/GCE clusters are granted "compute/rw" permissions. If you setup a cluster without these permissions, GLBC is useless and you should delete the controller as described in the [section below](#disabling-glbc). If you don't delete the controller it will keep restarting.
* [Default backends](https://cloud.google.com/compute/docs/load-balancing/http/url-map#url_map_simplest_case): All L7 Loadbalancers created by GLBC have a default backend. If you don't specify one in your Ingress, GLBC will assign the 404 default backend mentioned above.
* [Teardown](README.md#deletion): The recommended way to tear down a cluster with active Ingresses is to either delete each Ingress, or hit the `/delete-all-and-quit` endpoint on GLBC, before invoking a cluster teardown script (eg: kube-down.sh). You will have to manually cleanup GCE resources through the [cloud console](https://cloud.google.com/compute/docs/console#access) or [gcloud CLI](https://cloud.google.com/compute/docs/gcloud-compute/) if you simply tear down the cluster with active Ingresses.

## Prerequisites

Before you can receive traffic through the GCE L7 Loadbalancer Controller you need:
* A Working Kubernetes cluster >= 1.1
* At least 1 Kubernetes [NodePort Service](../../../../docs/user-guide/services.md#type-nodeport) (this is the endpoint for your Ingress)
* A single instance of the L7 Loadbalancer Controller pod (if you're using the default GCE setup, this should already be running in the `kube-system` namespace)

## Quota

GLBC is not aware of your GCE quota. As of this writing users get 3 [GCE Backend Services](https://cloud.google.com/compute/docs/load-balancing/http/backend-service) by default. If you plan on creating Ingresses for multiple Kubernetes Services, remember that each one requires a backend service, and request quota. Should you fail to do so the controller will poll periodically and grab the first free backend service slot it finds. You can view your quota:

```console
$ gcloud compute project-info describe --project myproject
```
See [GCE documentation](https://cloud.google.com/compute/docs/resource-quotas#checking_your_quota) for how to request more.

## Latency

It takes ~1m to spin up a loadbalancer (this includes acquiring the public ip), and ~5-6m before the GCE api starts healthchecking backends. So as far as latency goes, here's what to expect:

Assume one creates the following simple Ingress:
```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress
spec:
  backend:
    # This will just loopback to the default backend of GLBC
    serviceName: default-http-backend
    servicePort: 80
```

* time, t=0
```console
$ kubectl get ing
NAME           RULE      BACKEND                   ADDRESS
test-ingress   -         default-http-backend:80
$ kubectl describe ing
No events.
```

* time, t=1m
```console
$ kubectl get ing
NAME           RULE      BACKEND                   ADDRESS
test-ingress   -         default-http-backend:80   130.211.5.27

$ kubectl describe ing
target-proxy:		k8s-tp-default-test-ingress
url-map:		    k8s-um-default-test-ingress
backends:		    {"k8s-be-32342":"UNKNOWN"}
forwarding-rule:	k8s-fw-default-test-ingress
Events:
  FirstSeen	LastSeen	Count	From				SubobjectPath	Reason	Message
  ─────────	────────	─────	────				─────────────	──────	───────
  46s		46s		1	{loadbalancer-controller }	Success	Created loadbalancer 130.211.5.27
```

* time, t=5m
```console
$ kubectl describe ing
target-proxy:		k8s-tp-default-test-ingress
url-map:		    k8s-um-default-test-ingress
backends:		    {"k8s-be-32342":"HEALTHY"}
forwarding-rule:	k8s-fw-default-test-ingress
Events:
  FirstSeen	LastSeen	Count	From				SubobjectPath	Reason	Message
  ─────────	────────	─────	────				─────────────	──────	───────
  46s		46s		1	{loadbalancer-controller }	Success	Created loadbalancer 130.211.5.27

```

## Health checks

Currently, all service backends must respond with a 200 on '/'. The content does not matter. If they fail to do so they will be deemed unhealthy by the GCE L7. This limitation is because there are 2 sets of health checks:
* From the kubernetes endpoints, taking the form of liveness/readiness probes
* From the GCE L7, which periodically pings '/'
We really want (1) to control the health of an instance but (2) is a GCE requirement. Ideally, we would point (2) at (1), but we still need (2) for pods that don't have a defined health check. This will probably get resolved when Ingress grows up.


## Running multiple loadbalanced clusters in the same GCE project

If you're creating multiple clusters that will use Ingress within a single GCE project, you MUST assign a UID to GLBC so it doesn't stomp on resources from another cluster. You can do so by:
```console
$ kubectl get rc --namespace=kube-system
NAME                             DESIRED   CURRENT   AGE
elasticsearch-logging-v1         2         2         26m
heapster-v1.0.0                  1         1         26m
kibana-logging-v1                1         1         26m
kube-dns-v11                     1         1         26m
kubernetes-dashboard-v1.0.0      1         1         26m
l7-lb-controller-v0.6.0          1         1         26m
monitoring-influxdb-grafana-v3   1         1         26m

$ kubectl edit rc l7-lb-controller-v0.6.0 --namespace=kube-system
```

And modify the args passed to the controller:
```yaml
      - args:
        - --default-backend-service=kube-system/default-http-backend
        - --sync-period=300s
        - --cluster-uid=uid
```

Saving the file should update the RC but not the existing pod. To do so, just delete the pod, and the RC will create a new one with the --cluster-uid args.
```console
$ kubectl delete pod -l name=glbc --namespace=kube-system
pod "l7-lb-controller-v0.6.0-ud9ix" deleted
$ kubectl get pod --namespace=kube-system -l name=glbc -o yaml | grep cluster-uid
      - --cluster-uid=uid
```

## Creating the firewall rule for GLBC health checks

A default GKE/GCE cluster needs at least 1 firewall rule for GLBC to function. You can create it thus:
```console
$ gcloud compute firewall-rules create allow-130-211-0-0-22 \
  --source-ranges 130.211.0.0/22 \
  --target-tags $TAG \
  --allow tcp:$NODE_PORT
```

Where `130.211.0.0/22` is the source range of the GCE L7, `$NODE_PORT` is the node port your Service is exposed on, i.e:
```console
$ kubectl get -o jsonpath="{.spec.ports[0].nodePort}" services ${SERVICE_NAME}
```

and `$TAG` is an optional list of GKE instance tags, i.e:
```console
$ kubectl get nodes | awk '{print $1}' | tail -n +2 | grep -Po 'gke-[0-9,a-z]+-[0-9,a-z]+-node' | uniq
```

## Static and Ephemeral IPs

GCE has a concept of [ephemeral](https://cloud.google.com/compute/docs/instances-and-network#ephemeraladdress) and [static](https://cloud.google.com/compute/docs/instances-and-network#reservedaddress) IPs. A production website would always want a static IP, which ephemeral IPs are cheaper (both in terms of quota and cost), and are therefore better suited for experimentation.
* Creating a HTTP Ingress (i.e an Ingress without a TLS section) allocates an ephemeral IP, because we don't believe HTTP is the right way to deploy an app.
* Creating an Ingress with a TLS section allocates a static IP, because GLBC assumes you mean business.
* Modifying an Ingress and adding a TLS section allocates a static IP, but the IP *will* change. This is a beta limitation.
* You can [promote](https://cloud.google.com/compute/docs/instances-and-network#promote_ephemeral_ip) an ephemeral to a static IP by hand, if required.

## Disabling GLBC

Since GLBC runs as a cluster addon, you cannot simply delete the RC. The easiest way to disable it is to do as follows:

* IFF you want to tear down existing L7 loadbalancers, hit the /delete-all-and-quit endpoint on the pod:

```console
$ kubectl get pods --namespace=kube-system
NAME                                               READY     STATUS    RESTARTS   AGE
l7-lb-controller-7bb21                             1/1       Running   0          1h
$ kubectl exec l7-lb-controller-7bb21 -c l7-lb-controller curl http://localhost:8081/delete-all-and-quit --namespace=kube-system
$ kubectl logs l7-lb-controller-7b221 -c l7-lb-controller --follow
...
I1007 00:30:00.322528       1 main.go:160] Handled quit, awaiting pod deletion.
```

* Nullify the RC (but don't delete it or the addon controller will "fix" it for you)
```console
$ kubectl scale rc l7-lb-controller --replicas=0 --namespace=kube-system
```


