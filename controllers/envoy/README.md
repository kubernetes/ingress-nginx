# Envoy ingress controller

Runs a kubernetes ingress controller mapping the ingresses to lyft/envoy.

The k8s ingress gives a list of all services, pods, and "ingress" definitions for hte cluster. These are transformed into the format needed t configure `envoy`, presenting them as a [LDS](https://lyft.github.io/envoy/docs/configuration/listeners/lds.html), [RDS](https://lyft.github.io/envoy/docs/configuration/http_conn_man/rds.html), [SDS](https://lyft.github.io/envoy/docs/configuration/cluster_manager/sds.html), and [CDS](https://lyft.github.io/envoy/docs/configuration/cluster_manager/cds.html) for `envoy.

Once an initial sync of all the ingresses has completed, a instance of `envoy` is started locally. a initial json configuration with all routes will be created and written to disk. Envoy will then be started using that static configuration file. The config will have the envoy instance watch the discovery service with a very short ping interval (Ideally would rather push to envoy than be polled for "Fastest" updates, but 1-5s polling should be good enough).
