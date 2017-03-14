# Static IPs

This example demonstrates how to assign a [static-ip](https://cloud.google.com/compute/docs/configure-instance-ip-addresses#reserve_new_static) to an Ingress on GCE.

## Prerequisites

You need a [TLS cert](/examples/PREREQUISITES.md#tls-certificates) and a [test HTTP service](/examples/PREREQUISITES.md#test-http-service) for this example.
You will also need to make sure you Ingress targets exactly one Ingress
controller by specifying the [ingress.class annotation](/examples/PREREQUISITES.md#ingress-class),
and that you have an ingress controller [running](/examples/deployment) in your cluster.

## Acquiring a static IP

In GCE, static IP belongs to a given project until the owner decides to release
it. If you create a static IP and assign it to an Ingress, deleting the Ingress
or tearing down the GKE cluster *will not* delete the static IP. You can check
the static IPs you have as follows

```console
$ gcloud compute addresses list --global
NAME                     REGION  ADDRESS          STATUS
test-ip                          35.186.221.137   RESERVED

$ gcloud compute addresses list
NAME                      REGION       ADDRESS          STATUS
test-ip                                35.186.221.137   RESERVED
test-ip                   us-central1  35.184.21.228    RESERVED
```

Note the difference between a regional and a global static ip. Only global
static-ips will work with Ingress. If you don't already have an IP, you can
create it

```console
$ gcloud compute addresses create test-ip --global
Created [https://www.googleapis.com/compute/v1/projects/kubernetesdev/global/addresses/test-ip].
---
address: 35.186.221.137
creationTimestamp: '2017-01-31T10:32:29.889-08:00'
description: ''
id: '9221457935391876818'
kind: compute#address
name: test-ip
selfLink: https://www.googleapis.com/compute/v1/projects/kubernetesdev/global/addresses/test-ip
status: RESERVED
```

## Assigning a static IP to an Ingress

You can now add the static IP from the previous step to an Ingress,
by specifying the `kubernetes.io/global-static-ip-name` annotation,
the example yaml in this directory already has it set to `test-ip`

```console
$ kubectl create -f gce-static-ip-ingress.yaml
ingress "static-ip" created

$ gcloud compute addresses list test-ip
NAME     REGION       ADDRESS         STATUS
test-ip               35.186.221.137  IN_USE
test-ip  us-central1  35.184.21.228   RESERVED

$ kubectl get ing
NAME        HOSTS     ADDRESS          PORTS     AGE
static-ip   *         35.186.221.137   80, 443   1m

$ curl 35.186.221.137 -Lk
CLIENT VALUES:
client_address=10.180.1.1
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://35.186.221.137:8080/
...
```

## Retaining the static IP

You can test retention by deleting the Ingress

```console
$ kubectl delete -f gce-static-ip-ingress.yaml
ingress "static-ip" deleted

$ kubectl get ing
No resources found.

$ gcloud compute addresses list test-ip --global
NAME     REGION       ADDRESS         STATUS
test-ip               35.186.221.137  RESERVED
```

## Promote ephemeral to static IP

If you simply create a HTTP Ingress resource, it gets an ephemeral IP

```console
$ kubectl create -f gce-http-ingress.yaml
ingress "http-ingress" created

$ kubectl get ing
NAME           HOSTS     ADDRESS         PORTS     AGE
http-ingress   *         35.186.195.33   80        1h

$ gcloud compute forwarding-rules list
NAME                                           REGION       IP_ADDRESS      IP_PROTOCOL  TARGET
k8s-fw-default-http-ingress--32658fa96c080068               35.186.195.33   TCP          k8s-tp-default-http-ingress--32658fa96c080068
```

Note that because this is an ephemeral IP, it won't show up in the output of
`gcloud compute addresses list`.

If you either directly create an Ingress with a TLS section, or modify a HTTP
Ingress to have a TLS section, it gets a static IP.

```console
$ kubectl patch ing http-ingress -p '{"spec":{"tls":[{"secretName":"tls-secret"}]}}'
"http-ingress" patched

$ kubectl get ing
NAME           HOSTS     ADDRESS         PORTS     AGE
http-ingress   *         35.186.195.33   80, 443   1h

$ gcloud compute addresses list
NAME                                           REGION       ADDRESS          STATUS
k8s-fw-default-http-ingress--32658fa96c080068               35.186.195.33    IN_USE
```

