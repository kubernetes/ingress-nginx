# Ingress FAQ

This page contains general FAQ for Ingress, there is also a per-backend FAQ
in this directory with site specific information.

Table of Contents
=================

* [How is Ingress different from a Service?](#how-is-ingress-different-from-a-service)
* [I created an Ingress and nothing happens, what now?](#i-created-an-ingress-and-nothing-happens-what-now)
* [How do I deploy an Ingress controller?](#how-do-i-deploy-an-ingress-controller)
* [Are Ingress controllers namespaced?](#are-ingress-controllers-namespaced)
* [How do I disable an Ingress controller?](#how-do-i-disable-an-ingress-controller)
* [How do I run multiple Ingress controllers in the same cluster?](#how-do-i-run-multiple-ingress-controllers-in-the-same-cluster)
* [How do I contribute a backend to the generic Ingress controller?](#how-do-i-contribute-a-backend-to-the-generic-ingress-controller)
* [Is there a catalog of existing Ingress controllers?](#is-there-a-catalog-of-existing-ingress-controllers)
* [How are the Ingress controllers tested?](#how-are-the-ingress-controllers-tested)
* [An Ingress controller E2E is failing, what should I do?](#an-ingress-controller-e2e-is-failing-what-should-i-do)
* [Is there a roadmap for Ingress features?](#is-there-a-roadmap-for-ingress-features)

## How is Ingress different from a Service?

The Kubernetes Service is an abstraction over endpoints (pod-ip:port pairings).
The Ingress is an abstraction over Services. This doesn't mean all Ingress
controller must route *through* a Service, but rather, that routing, security
and auth configuration is represented in the Ingress resource per Service, and
not per pod. As long as this configuration is respected, a given Ingress
controller is free to route to the DNS name of a Service, the VIP, a NodePort,
or directly to the Service's endpoints.

## I created an Ingress and nothing happens, what now?

Run `describe` on the Ingress. If you see create/add events, you have an Ingress
controller running in the cluster, otherwise, you either need to deploy or
restart your Ingress controller. If the events associated with an Ingress are
insufficient to debug, consult the controller specific FAQ.

## How do I deploy an Ingress controller?

The following platforms currently deploy an Ingress controller addon: GCE, GKE,
minikube. If you're running on any other platform, you can deploy an Ingress
controller by following [this](/examples/deployment) example.

## Are Ingress controllers namespaced?

Ingress is namespaced, this means 2 Ingress objects can have the same name in 2
namespaces, and must only point to Services in its own namespace. An admin can
deploy an Ingress controller such that it only satisfies Ingress from a given
namespace, but by default, controllers will watch the entire Kubernetes cluster
for unsatisfied Ingress.

## How do I disable an Ingress controller?

Either shutdown the controller satisfying the Ingress, or use the
`Ingress-class` annotation, as follows:

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

Setting the annotation to any value other than "gce" or the empty string, will
force the GCE controller to ignore your Ingress. The same applies for the nginx
controller.

To completely stop the Ingress controller on GCE/GKE, please see [this](gce.md#host-do-i-disable-the-ingress-controller) faq.

## How do I run multiple Ingress controllers in the same cluster?

Multiple Ingress controllers can co-exist and key off the `ingress-class`
annotation, as shown in this [faq](#how-do-i-run-multiple-ingress-controllers-in-the-same-cluster),
as well as in [this](/examples/daemonset/nginx) example.

## How do I contribute a backend to the generic Ingress controller?

First check the [catalog](#is-there-a-catalog-of-existing-ingress-controllers), to make sure you really need to write one.

1. Write a [generic backend](/examples/custom-controller)
2. Keep it in your own repo, make sure it passes the [conformance suite](https://github.com/kubernetes/kubernetes/blob/master/test/e2e/ingress_utils.go#L112)
3. Submit an example(s) in the appropriate subdirectories [here](/examples/README.md)
4. Add it to the catalog

## Is there a catalog of existing Ingress controllers?

Yes, a non-comprehensive [catalog](/docs/catalog.md) exists.

## How are the Ingress controllers tested?

Testing for the Ingress controllers is divided between:
* Ingress repo: unit tests and pre-submit integration tests run via travis
* Kubernetes repo: [pre-submit e2e](https://k8s-testgrid.appspot.com/google-gce#gce&include-filter-by-regex=Loadbalancing),
  [post-merge e2e](https://k8s-testgrid.appspot.com/google-gce#gci-gce-ingress),
  [per release-branch e2e](https://k8s-testgrid.appspot.com/google-gce#gci-gce-ingress-1.5)

The configuration for jenkins e2e tests are located [here](https://github.com/kubernetes/test-infra).
The Ingress E2Es are located [here](https://github.com/kubernetes/kubernetes/blob/master/test/e2e/network/ingress.go),
each controller added to that suite must consistently pass the [conformance suite](https://github.com/kubernetes/kubernetes/blob/master/test/e2e/framework/ingress_utils.go#L129).

## An Ingress controller E2E is failing, what should I do?

First, identify the reason for failure.

* Look at the build log, if there's nothing obvious, search for quota issues.
  * Find events logged by the controller in the build log
  * Ctrl+f "quota" in the build log
* If the failure is in the GCE controller:
  * Navigate to the test artifacts for that run and look at glbc.log, [eg](http://gcsweb.k8s.io/gcs/kubernetes-jenkins/logs/ci-kubernetes-e2e-gci-gce-ingress-release-1.5/1234/artifacts/bootstrap-e2e-master/)
  * Look up the `PROJECT=` line in the build log, and navigate to that project
    looking for quota issues (`gcloud compute project-info describe project-name`
    or navigate to the cloud console > compute > quotas)
* If the failure is for a non-cloud controller (eg: nginx)
  * Make sure the firewall rules required by the controller are opened on the
    right ports (80/443), since the jenkins builders run *outside* the
    Kubernetes cluster.

Note that you currently need help from a test-infra maintainer to access the GCE
test project. If you think the failures are related to project quota, cleanup
leaked resources and bump up quota before debugging the leak.

If the preceding identification process fails, it's likely that the Ingress api
is broken upstream. Try to setup a [dev environment](/docs/dev/setup.md) from
HEAD and create an Ingress. You should be deploying the [latest](https://github.com/kubernetes/ingress/releases)
release image to the local cluster.

If neither of these 2 strategies produces anything useful, you can either start
reverting images, or digging into the underlying infrastructure the e2es are
running on for more nefarious issues (like permission and scope changes for
some set of nodes on which an Ingress controller is running).

## Is there a roadmap for Ingress features?

The community is working on it. There are currently too many efforts in flight
to serialize into a flat roadmap. You might be interested in the following issues:
* Loadbalancing [umbrella issue](https://github.com/kubernetes/kubernetes/issues/24145)
* Service proxy [proposal](https://groups.google.com/forum/#!topic/kubernetes-sig-network/weni52UMrI8)
* Better [routing rules](https://github.com/kubernetes/kubernetes/issues/28443)
* Ingress [classes](https://github.com/kubernetes/kubernetes/issues/30151)

As well as the issues in this repo.

