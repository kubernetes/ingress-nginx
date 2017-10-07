# NGINX Ingress running in AWS

This example shows how is possible to use the nginx ingress controller in AWS behind an ELB configured with Proxy Protocol.

```console
kubectl create -f ./nginx-ingress-controller.yaml
```

This command creates:
- a default backend deployment and service.
- a service with `type: LoadBalancer` configuring Proxy Protocol in the ELB (`service.beta.kubernetes.io/aws-load-balancer-proxy-protocol: '*'`).
- a configmap for the ingress controller enabling proxy protocol in NGINX (`use-proxy-protocol: "true"`)
- a deployment for the ingress controller

Is the proxy protocol necessary?

No but only enabling the protocol is possible to keep the real source IP address requesting the connection.

### References

- http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-proxy-protocol.html
- https://www.nginx.com/resources/admin-guide/proxy-protocol/
