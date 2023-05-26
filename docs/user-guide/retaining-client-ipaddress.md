
## Retaining Client IPAddress

Please read this https://kubernetes.github.io/ingress-nginx/user-guide/miscellaneous/#source-ip-address , to get details of retaining the client IPAddress.

### Using proxy-protocol

Please read this https://kubernetes.github.io/ingress-nginx/user-guide/miscellaneous/#proxy-protocol , to use proxy-protocol for retaining client IPAddress


### Using the K8S spec service.spec.externalTrafficPolicy

```
% kubectl explain service.spec.externalTrafficPolicy
KIND:       Service
VERSION:    v1

FIELD: externalTrafficPolicy <string>

DESCRIPTION:
    externalTrafficPolicy describes how nodes distribute service traffic they
    receive on one of the Service's "externally-facing" addresses (NodePorts,
    ExternalIPs, and LoadBalancer IPs). If set to "Local", the proxy will
    configure the service in a way that assumes that external load balancers
    will take care of balancing the service traffic between nodes, and so each
    node will deliver traffic only to the node-local endpoints of the service,
    without masquerading the client source IP. (Traffic mistakenly sent to a
    node with no endpoints will be dropped.) The default value, "Cluster", uses
    the standard behavior of routing to all endpoints evenly (possibly modified
    by topology and other features). Note that traffic sent to an External IP or
    LoadBalancer IP from within the cluster will always get "Cluster" semantics,
    but clients sending to a NodePort from within the cluster may need to take
    traffic policy into account when picking a node.
    
    Possible enum values:
     - `"Cluster"` routes traffic to all endpoints.
     - `"Local"` preserves the source IP of the traffic by routing only to
    endpoints on the same node as the traffic was received on (dropping the
    traffic if there are no local endpoints).

```


- Setting the field `externalTrafficPolicy`, in the ingress-controller service, to a value of `Local` retains the client's ipaddress, within the scope explained above
