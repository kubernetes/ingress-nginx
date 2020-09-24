---
title: Remove static SSL configuration mode
authors:
  - "@aledbf"
reviewers:
  - "@ElvinEfendi"
approvers:
  - "@ElvinEfendi"
editor: TBD
creation-date: 2019-07-24
last-updated: 2019-07-24
status: implementable
see-also:
replaces:
superseded-by:
---

#  Remove static SSL configuration mode

## Table of Contents

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [Implementation Details/Notes/Constraints](#implementation-detailsnotesconstraints)
- [Drawbacks](#drawbacks)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

Since release [0.19.0](https://github.com/kubernetes/ingress-nginx/releases/tag/nginx-0.19.0) is possible to configure SSL certificates without the need of NGINX reloads (thanks to lua) and after release [0.24.0](https://github.com/kubernetes/ingress-nginx/releases/tag/nginx-0.24.0) the default enabled mode is dynamic.

## Motivation

The static configuration implies reloads, something that affects the majority of the users.

### Goals

- Deprecation of the flag `--enable-dynamic-certificates`.
- Cleanup of the codebase.

### Non-Goals

- Features related to certificate authentication are not changed in any way.

## Proposal

- Remove static SSL configuration

### Implementation Details/Notes/Constraints

- Deprecate the flag Move the directives `ssl_certificate` and `ssl_certificate_key` from each server block to the `http` section. These settings are required to avoid NGINX errors in the logs.
- Remove any action of the flag `--enable-dynamic-certificates`

## Drawbacks

## Alternatives

Keep both implementations
