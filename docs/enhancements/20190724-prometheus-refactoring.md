---
title: Refactoring of Prometheus metrics
authors:
  - "@aledbf"
reviewers:
  - "@ElvinEfendi"
approvers:
  - "@ElvinEfendi"
editor: TBD
creation-date: 2019-07-24
last-updated: 2019-07-24
status: provisional
see-also:
replaces:
superseded-by:
---

# Refactoring of Prometheus metrics

## Table of Contents

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [Implementation Details/Notes/Constraints](#implementation-detailsnotesconstraints)
  - [Info metrics](#info-metrics)
    - [<em>Controller</em>](#controller)
    - [<em>Ingresses</em> (only the leader handle this metrics)](#ingresses-only-the-leader-handle-this-metrics)
  - [Metrics](#metrics)
- [Drawbacks](#drawbacks)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

Ingress-nginx currently provides prometheus metrics using Lua to extract and dispatch the data from NGINX to the ingress controller during the [log phase](http://nginx.org/en/docs/dev/development_guide.html#http_phases) making an HTTP request to a Unix socket. This socket is an HTTP server exposed by the ingress controller where the data received is used to update the defined prometheus metric. The metrics are exposed using the endpoint `/metrics` in the ingress controller.

## Motivation

The current implementation lacks enough configuration options to allow the customization of the features to use in the metrics. Also, due to the lack of iterations to improve metrics, the current state has several issues:

- Lack of details about endpoints
- Performance issues in scenarios with high RPS
- Use of histogram and summaries introduce high cardinality
- High cardinality of metrics in general.

### Goals

- Customizable prometheus metrics for any workload without compromising NGINX performance.
- Metrics counter are initialized at start/reload [#4066](https://github.com/kubernetes/ingress-nginx/issues/4066)
- Removal of flags, moving these flags to the configuration configmap. The reason for this is to not require a change in the ingress controller deployment.
- Expose number of endpoints per ingress via prometheus [#3899](https://github.com/kubernetes/ingress-nginx/issues/3899)
- Fix several issues:

  - https://github.com/kubernetes/ingress-nginx/issues/3936
  - https://github.com/kubernetes/ingress-nginx/issues/4026
  - https://github.com/kubernetes/ingress-nginx/issues/3645
  - https://github.com/kubernetes/ingress-nginx/issues/3818
  - https://github.com/kubernetes/ingress-nginx/issues/3713
  - https://github.com/kubernetes/ingress-nginx/pull/4139
  - https://github.com/kubernetes/ingress-nginx/issues/2924
  - https://github.com/kubernetes/ingress-nginx/issues/3898
  - https://github.com/kubernetes/ingress-nginx/pull/4139

### Non-Goals

Update of the Grafana dashboard [provided by the project](https://github.com/kubernetes/ingress-nginx/tree/master/deploy/grafana/dashboards). This should be done after the implementation of this KEP.

## Proposal

Test if using message pack we can improve the performance of the Lua communication with the ingress controller ([lua library](https://luarocks.org/modules/antirez/lua-cmsgpack))

One of the issues we have is the range of options and configuration users expected from the exposed metrics. This creates a maintenance problem due to the fact the metrics are hardcoded in the source code.
For this reason, one of the goals in this KEP is to introduce a CRD for metrics, allowing users to define which rules they want to use and which labels should be used.
As an example:

```yaml
kind: metric
metadata:
  name: request_duration_seconds
  namespace: default
spec:
  help: The request processing time in milliseconds
  constLabels:
    - controller_pod
  labels:
    - ingress_uid # label to join with the ingress info metric using group_left
    - path        # same ^^
    - method
    - status
  type: counter|histogram|summary|gauge
  buckets:
    type: linear|exponential
    values:
      - 10
      - 10
      - 10
```

The default deployment should provide the same metrics we have now using this new `metric` type, using the configuration configmap to set the list of metrics to use, allowing users to define new types
This approach allows us to also provide other types defined in other projects like VTS module or prometheus collector from nginx inc https://github.com/nginxinc/nginx-prometheus-exporter

Use info metrics to reduce the number of labels (even constants) in the metrics. This will decrement the size of the metrics and just doing one join we can get the same cardinality.

### Implementation Details/Notes/Constraints

### Info metrics

#### *Controller*

- controller_class
- controller_namespace
- controller_pod

#### *Ingresses* (only the leader handle this metrics)

|Field name|Description|
| --- | --- |
|uid| Unique identifier of the ingress. This field is the key for the joins|
|name| Name of the ingress|
|namespace| Namespace where the ingress and the service are located|
|class| Value of the annoation kubernetes.io/ingress.class|
|host| Hostname of the ingress|
|path| Path of the `http` backend in the ingress|
|service| Name of the service being exposed|

### Metrics

**Name:** nginx_ingress_controller_bytes_sent

**Description:** The the number of bytes sent to a client

**Type:** histogram

**Labels:**

- ingress_uid
- method
- status

----------

**Name:** nginx_ingress_controller_config_hash 

**Description:** Running configuration hash actually running

**Type:** gauge

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_config_last_reload_successful

**Description:** Whether the last configuration reload attempt was successful

**Type:** gauge

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_config_last_reload_successful_timestamp_seconds

**Description:** Timestamp of the last successful configuration reload

**Type:** gauge

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_ingress_upstream_latency_seconds

**Description:** Upstream service latency per Ingress

**Type:** summary

**Labels:**

- controller_pod
- ingress_uid

----------

**Name:** nginx_ingress_controller_nginx_process_connections

**Description:** Current number of client connections with state {reading, writing, waiting}

**Type:** gauge

**Labels:**

- controller_pod
- state (reading, waiting, writing)

----------

**Name:** nginx_ingress_controller_nginx_process_connections_total

**Description:** Total number of connections with state {active, accepted, handled}

**Type:** counter

**Labels:**

- controller_pod
- state (accepted, active, handled)

----------

**Name:** nginx_ingress_controller_nginx_process_cpu_seconds_total

**Description:** CPU usage in seconds

**Type:** counter

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_nginx_process_num_procs

**Description:** Number of processes

**Type:** gauge

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_nginx_process_oldest_start_time_seconds

**Description:** Start time in seconds since 1970/01/01

**Type:** gauge

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_nginx_process_read_bytes_total

**Description:** Number of bytes read

**Type:** counter

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_nginx_process_requests_total

**Description:** Total number of client requests

**Type:** counter

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_nginx_process_resident_memory_bytes

**Description:** Number of bytes of memory in use

**Type:** gauge

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_nginx_process_virtual_memory_bytes

**Description:** Number of bytes of memory in use

**Type:** gauge

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_nginx_process_write_bytes_total

**Description:** Number of bytes written

**Type:** counter

**Labels:**

- controller_pod

----------

**Name:** nginx_ingress_controller_request_duration_seconds

**Description:** The request processing time in milliseconds

**Type:** histogram

**Labels:**

- controller_pod
- ingress_uid
- method
- status

----------

**Name:** nginx_ingress_controller_request_size

**Description:** The request length (including request line, header, and request body)

**Type:** histogram

**Labels:**

- controller_pod
- ingress_uid
- method
- status

----------

**Name:** nginx_ingress_controller_requests

**Description:** The total number of client requests.

**Type:** counter

**Labels:**

- controller_pod
- ingress_uid
- status

----------

TODO: endpoint metrics

## Drawbacks

Any change to the metrics will be a breaking change and affect any automation build on top of it.

## Alternatives

There are no real alternatives to the existing way to extract metrics from NGINX due to the lack of native support 
for metrics in the Open-source version and the lack of support of Lua in the [vts](https://github.com/vozlt/nginx-module-vts) module.
