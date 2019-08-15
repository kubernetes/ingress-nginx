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
last-updated: 2019-08-15
status: provisional
---

# Availability zone aware routing

## Table of Contents

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
<!-- /toc -->

## Summary

Teach ingress-nginx about availability zones where endpoints are running in. This way ingress-nginx pod will do its best to proxy to zone-local endpoint.

## Motivation

When users run their services across multiple availability zones they usually pay for egress traffic between zones. Providers such as GCP, AWS charges money for that.
ingress-nginx when picking an endpoint to route request to does not consider whether the endpoint is in different zone or the same one. That means it's at least equally likely
that it will pick an endpoint from another zone and proxy the request to it. In this situation response from the endpoint to ingress-nginx pod is considered as
inter zone traffic and costs money.


At the time of this writing GCP charges $0.01 per GB of inter zone egress traffic according to https://cloud.google.com/compute/network-pricing.
According to https://datapath.io/resources/blog/what-are-aws-data-transfer-costs-and-how-to-minimize-them/ AWS also charges the same amount of money sa GCP for cross zone, egress traffic.

This can be a lot of money depending on once's traffic. By teaching ingress-nginx about zones we can eliminate or at least decrease this cost.

Arguably inter-zone network latency should also be better than cross zone.

### Goals

* Given a regional cluster running ingress-nginx, ingress-nginx should do best effort to pick zone-local endpoint when proxying

### Non-Goals

* -
