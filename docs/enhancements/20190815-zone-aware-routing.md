---
title: Availability zone aware routing
authors:
  - "@ElvinEfendi"
reviewers:
  - "@aledbf"
approvers:
  - "@aledbf"
editor: TBD
creation-date: 2019-08-15
last-updated: 2019-08-16
status: implementable
---

# Availability zone aware routing

## Table of Contents

<!-- toc -->
- [Availability zone aware routing](#availability-zone-aware-routing)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
  - [Implementation History](#implementation-history)
  - [Drawbacks [optional]](#drawbacks-optional)
<!-- /toc -->

## Summary

Teach ingress-nginx about availability zones where endpoints are running in. This way ingress-nginx pod will do its best to proxy to zone-local endpoint.

## Motivation

When users run their services across multiple availability zones they usually pay for egress traffic between zones. Providers such as GCP, and Amazon EC2 usually charge extra for this feature.
ingress-nginx when picking an endpoint to route request to does not consider whether the endpoint is in a different zone or the same one. That means it's at least equally likely
that it will pick an endpoint from another zone and proxy the request to it. In this situation response from the endpoint to the ingress-nginx pod is considered
inter-zone traffic and usually costs extra money.


At the time of this writing, GCP charges $0.01 per GB of inter-zone egress traffic according to https://cloud.google.com/compute/network-pricing.
According to [https://datapath.io/resources/blog/what-are-aws-data-transfer-costs-and-how-to-minimize-them/](https://web.archive.org/web/20201008160149/https://datapath.io/resources/blog/what-are-aws-data-transfer-costs-and-how-to-minimize-them/) Amazon also charges the same amount of money as GCP for cross-zone, egress traffic.

This can be a lot of money depending on once's traffic. By teaching ingress-nginx about zones we can eliminate or at least decrease this cost.

Arguably inter-zone network latency should also be better than cross-zone.

### Goals

* Given a regional cluster running ingress-nginx, ingress-nginx should do best-effort to pick a zone-local endpoint when proxying
* This should not impact canary feature
* ingress-nginx should be able to operate successfully if there are no zonal endpoints

### Non-Goals

* This feature inherently assumes that endpoints are distributed across zones in a way that they can handle all the traffic from ingress-nginx pod(s) in that zone
* This feature will be relying on https://kubernetes.io/docs/reference/kubernetes-api/labels-annotations-taints/#failure-domainbetakubernetesiozone, it is not this KEP's goal to support other cases

## Proposal

The idea here is to have the controller part of ingress-nginx
(1) detect what zone its current pod is running in and
(2) detect the zone for every endpoint it knows about.
After that, it will post that data as part of endpoints to Lua land.
When picking an endpoint, the Lua balancer will try to pick zone-local endpoint first and
if there is no zone-local endpoint then it will fall back to current behavior.

Initially, this feature should be optional since it is going to make it harder to reason about the load balancing and not everyone might want that.

**How does controller know what zone it runs in?**
We can have the pod spec pass the node name using downward API as an environment variable.
Upon startup, the controller can get node details from the API based on the node name.
Once the node details are obtained
we can extract the zone from the `failure-domain.beta.kubernetes.io/zone` annotation.
Then we can pass that value to Lua land through Nginx configuration
when loading `lua_ingress.lua` module in `init_by_lua` phase.

**How do we extract zones for endpoints?**
We can have the controller watch create and update events on nodes in the entire cluster and based on that keep the map of nodes to zones in the memory.
And when we generate endpoints list, we can access node name using `.subsets.addresses[i].nodeName`
and based on that fetch zone from the map in memory and store it as a field on the endpoint.
__This solution assumes `failure-domain.beta.kubernetes.io/zone`__ annotation does not change until the end of the node's life. Otherwise, we have to
watch update events as well on the nodes and that'll add even more overhead.

Alternatively, we can get the list of nodes only when there's no node in the memory for the given node name. This is probably a better solution
because then we would avoid watching for API changes on node resources. We can eagerly fetch all the nodes and build node name to zone mapping on start.
From there on,  it will sync during endpoint building in the main event loop if there's no existing entry for the node of an endpoint.
This means an extra API call in case cluster has expanded.

**How do we make sure we do our best to choose zone-local endpoint?**
This will be done on the Lua side. For every backend, we will initialize two balancer instances:
(1) with all endpoints
(2) with all endpoints corresponding to the current zone for the backend.
Then given the request once we choose what backend
needs to serve the request, we will first try to use a zonal balancer for that backend.
If a zonal balancer does not exist (i.e. there's no zonal endpoint)
then we will use a general balancer.
In case of zonal outages, we assume that the readiness probe will fail and the controller will
see no endpoints for the backend and therefore we will use a general balancer.

We can enable the feature using a configmap setting. Doing it this way makes it easier to rollback in case of a problem.

## Implementation History

- initial version of KEP is shipped
- proposal and implementation details are done

## Drawbacks [optional]

More load on the Kubernetes API server.
