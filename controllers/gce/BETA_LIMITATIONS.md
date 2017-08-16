# GLBC: Beta limitations

As of the Kubernetes 1.7 release, the GCE L7 Loadbalancer controller is still a *beta* product.

This is a list of beta limitations:

* [IPs](#static-and-ephemeral-ips): Creating a simple HTTP Ingress will allocate an ephemeral IP. Creating an Ingress with a TLS section will allocate a static IP.
* [Latency](#latency): GLBC is not built for performance. Creating many Ingresses at a time can overwhelm it. It won't fall over, but will take its own time to churn through the Ingress queue.
* [Quota](#quota): By default, GCE projects are granted a quota of 3 Backend Services. This is insufficient for most Kubernetes clusters.
* [Oauth scopes](https://cloud.google.com/compute/docs/authentication): By default GKE/GCE clusters are granted "compute/rw" permissions. If you setup a cluster without these permissions, GLBC is useless and you should delete the controller as described in the [section below](#disabling-glbc). If you don't delete the controller it will keep restarting.
* [Default backends](https://cloud.google.com/compute/docs/load-balancing/http/url-map#url_map_simplest_case): All L7 Loadbalancers created by GLBC have a default backend. If you don't specify one in your Ingress, GLBC will assign the 404 default backend mentioned above.
* [Load Balancing Algorithms](#load-balancing-algorithms): The ingress controller doesn't support fine grained control over loadbalancing algorithms yet.
* [Large clusters](#large-clusters): Ingress on GCE isn't supported on large (>1000 nodes), single-zone clusters.
* [Teardown](README.md#deletion): The recommended way to tear down a cluster with active Ingresses is to either delete each Ingress, or hit the `/delete-all-and-quit` endpoint on GLBC, before invoking a cluster teardown script (eg: kube-down.sh). You will have to manually cleanup GCE resources through the [cloud console](https://cloud.google.com/compute/docs/console#access) or [gcloud CLI](https://cloud.google.com/compute/docs/gcloud-compute/) if you simply tear down the cluster with active Ingresses.
* [Changing UIDs](#changing-the-cluster-uid): You can change the UID used as a suffix for all your GCE cloud resources, but this requires you to delete existing Ingresses first.
* [Cleaning up](#cleaning-up-cloud-resources): You can delete loadbalancers that older clusters might've leaked due to premature teardown through the GCE console.

## Prerequisites

Before you can receive traffic through the GCE L7 Loadbalancer Controller you need:
* A Working Kubernetes cluster >= 1.1
* At least 1 Kubernetes [NodePort Service](../../../../docs/user-guide/services.md#type-nodeport) (this is the endpoint for your Ingress)
* A single instance of the L7 Loadbalancer Controller pod, if you're running Kubernetes < 1.3 (the GCP ingress controller runs on the master in later versions)

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

## Static and Ephemeral IPs

GCE has a concept of [ephemeral](https://cloud.google.com/compute/docs/instances-and-network#ephemeraladdress) and [static](https://cloud.google.com/compute/docs/instances-and-network#reservedaddress) IPs. A production website would always want a static IP, which ephemeral IPs are cheaper (both in terms of quota and cost), and are therefore better suited for experimentation.
* Creating a HTTP Ingress (i.e an Ingress without a TLS section) allocates an ephemeral IP, because we don't believe HTTP is the right way to deploy an app.
* Creating an Ingress with a TLS section allocates a static IP, because GLBC assumes you mean business.
* Modifying an Ingress and adding a TLS section allocates a static IP, but the IP *will* change. This is a beta limitation.
* You can [promote](https://cloud.google.com/compute/docs/instances-and-network#promote_ephemeral_ip) an ephemeral to a static IP by hand, if required.

## Load Balancing Algorithms

Right now, a kube-proxy nodePort is a necessary condition for Ingress on GCP. This is because the cloud lb doesn't understand how to route directly to your pods. Incorporating kube-proxy and cloud lb algorithms so they cooperate toward a common goal is still a work in progress. If you really want fine grained control over the algorithm, you should deploy the nginx ingress controller.

## Large clusters

Ingress is not yet supported on single zone clusters of size > 1000 nodes ([issue](https://github.com/kubernetes/contrib/issues/1724)). If you'd like to use Ingress on a large cluster, spread it across 2 or more zones such that no single zone contains more than a 1000 nodes. This is because there is a [limit](https://cloud.google.com/compute/docs/instance-groups/creating-groups-of-managed-instances) to the number of instances one can add to a single GCE Instance Group. In a multi-zone cluster, each zone gets its own instance group.

## Disabling GLBC

Setting the annotation `kubernetes.io/ingress.class` to any value other than "gce" or the empty string, will force the GCE Ingress controller to ignore your Ingress. Do this if you wish to use one of the other Ingress controllers at the same time as the GCE controller, eg:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test
  annotations:
    kubernetes.io/ingress.class: "nginx"
spec:
  tls:
  - secretName: tls-secret
  backend:
    serviceName: echoheaders-https
    servicePort: 80
```

As of Kubernetes 1.3, GLBC runs as a static pod on the master. If you want to totally disable it, you can ssh into the master node and delete the GLBC manifest file found at `/etc/kubernetes/manifests/glbc.manifest`. You can also disable it on GKE at cluster bring-up time through the `disable-addons` flag, eg:

```console
gcloud container clusters create mycluster --network "default" --num-nodes 1 \
--machine-type n1-standard-2 --zone $ZONE \
--disable-addons HttpLoadBalancing \
--disk-size 50 --scopes storage-full
```

## Changing the cluster UID

The Ingress controller configures itself to add the UID it stores in a configmap in the `kube-system` namespace.

```console
$ kubectl --namespace=kube-system get configmaps
NAME          DATA      AGE
ingress-uid   1         12d

$ kubectl --namespace=kube-system get configmaps -o yaml
apiVersion: v1
items:
- apiVersion: v1
  data:
    uid: UID
  kind: ConfigMap
...
```

You can pick a different UID, but this requires you to:

1. Delete existing Ingresses
2. Edit the configmap using `kubectl edit`
3. Recreate the same Ingress

After step 3 the Ingress should come up using the new UID as the suffix of all cloud resources. You can't simply change the UID if you have existing Ingresses, because
renaming a cloud resource requires a delete/create cycle that the Ingress controller does not currently automate. Note that the UID in step 1 might be an empty string,
if you had a working Ingress before upgrading to Kubernetes 1.3.

__A note on setting the UID__: The Ingress controller uses the token `--` to split a machine generated prefix from the UID itself. If the user supplied UID is found to
contain `--` the controller will take the token after the last `--`, and use an empty string if it ends with `--`. For example, if you insert `foo--bar` as the UID,
the controller will assume `bar` is the UID. You can either edit the configmap and set the UID to `bar` to match the controller, or delete existing Ingresses as described
above, and reset it to a string bereft of `--`.

## Cleaning up cloud resources

If you deleted a GKE/GCE cluster without first deleting the associated Ingresses, the controller would not have deleted the associated cloud resources. If you find yourself in such a situation, you can delete the resources by hand:

1. Navigate to the [cloud console](https://console.cloud.google.com/) and click on the "Networking" tab, then choose "LoadBalancing"
2. Find the loadbalancer you'd like to delete, it should have a name formatted as: k8s-um-ns-name--UUID
3. Delete it, check the boxes to also cascade the deletion down to associated resources (eg: backend-services)
4. Switch to the "Compute Engine" tab, then choose "Instance Groups"
5. Delete the Instance Group allocated for the leaked Ingress, it should have a name formatted as: k8s-ig-UUID
