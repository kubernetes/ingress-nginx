# Exposing TCP and UDP services

Ingress does not support TCP or UDP services. For this reason this Ingress controller uses the flags `--tcp-services-configmap` and `--udp-services-configmap` to point to an existing config map where the key is the external port to use and the value indicates the service to expose using the format:
`<namespace/service name>:<service port>:[PROXY]:[PROXY]`

It is also possible to use a number or the name of the port. The two last fields are optional.
Adding `PROXY` in either or both of the two last fields we can use [Proxy Protocol](https://www.nginx.com/resources/admin-guide/proxy-protocol) decoding (listen) and/or encoding (proxy_pass) in a TCP service. 
The first `PROXY` controls the decode of the proxy protocol and the second `PROXY` controls the encoding using proxy protocol. 
This allows an incoming connection to be decoded or an outgoing connection to be encoded. It is also possible to arbitrate between two different proxies by turning on the decode and encode on a TCP service. 

The next example shows how to expose the service `example-go` running in the namespace `default` in the port `8080` using the port `9000`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tcp-services
  namespace: ingress-nginx
data:
  9000: "default/example-go:8080"
```

Since 1.9.13, NGINX provides [UDP Load Balancing](https://www.nginx.com/blog/announcing-udp-load-balancing/).
Similarly, the next example shows how to expose the service `kube-dns` running in the namespace `kube-system` in the port `53` using the port `53`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: udp-services
  namespace: ingress-nginx
data:
  53: "kube-system/kube-dns:53"
```

If TCP/UDP proxy support is used, then those ports need to be exposed in the Service defined for the Ingress. 

Notice that Kubernetes does not support mixed protocols for LoadBalancer Services yet. This means that, for each protocol, a dedicated LoadBalancer and Ingress Controller deployment are necessary. [This documentation](https://kubernetes.github.io/ingress-nginx/#how-to-easily-install-multiple-instances-of-the-ingress-nginx-controller-in-the-same-cluster) shows how to install multiple instances of Ingress NGINX Controllers in a K8s cluster. 

To create an Ingress NGINX Controller for the TCP services:
```bash
helm install ingress-nginx-tcp ingress-nginx/ingress-nginx  \
--namespace ingress-nginx \
--set controller.ingressClassResource.name=nginx-tcp \
--set controller.ingressClassResource.controllerValue="k8s.io/ingress-nginx-tcp" \
--set controller.ingressClassResource.enabled=true \
--set controller.ingressClassByName=true
```

To create an Ingress NGINX Controller for the UDP services:
```bash
helm install ingress-nginx-udp ingress-nginx/ingress-nginx  \
--namespace ingress-nginx \
--set controller.ingressClassResource.name=nginx-udp \
--set controller.ingressClassResource.controllerValue="k8s.io/ingress-nginx-udp" \
--set controller.ingressClassResource.enabled=true \
--set controller.ingressClassByName=true
```

Now, it is necessary to point the ConfigMaps created previously to each Ingress Controller.

In the TCP Controller, add the following under `spec.template.spec.containers.args`:
```yaml
- --tcp-services-configmap=$(POD_NAMESPACE)/tcp-services
```

And, in the UDP:
```yaml
- --udp-services-configmap=$(POD_NAMESPACE)/udp-services
```

One way of patching these deployments to point these ConfigMaps is using `kubectl`:
```bash
kubectl edit deployments -n kube-system ingress-nginx-controller
```

For example, the result for the UDP Controller Deployment should be something like:
```yaml
...
    spec:
      containers:
      - args:
        - /nginx-ingress-controller
        - --publish-service=kube-system/ingress-nginx-controller
        - --election-id=ingress-controller-leader
        - --ingress-class=nginx
        - --configmap=kube-system/ingress-nginx-controller
        - --udp-services-configmap=$(POD_NAMESPACE)/udp-services
        - --validating-webhook=:8443
        - --validating-webhook-certificate=/usr/local/certificates/cert
        - --validating-webhook-key=/usr/local/certificates/key
...
```

Next, patch each Ingress NGINX Controller Deployment so that it listen a certain port and can route traffic to its corresponding service.

For the UDP service, it would be like the following:
```yaml
spec:
  template:
    spec:
      containers:
        - name: controller
          ports:
            - containerPort: 53
              hostPort: 53
```
Create a file called `ingress-nginx-udp-deployment-patch.yaml` and paste the contents above.

To apply these changes, simply:
```bash
kubectl patch deployment ingress-nginx-udp-controller \ 
--patch "$(cat ingress-nginx-udp-controller-patch.yaml)" \
--namespace ingress-nginx
```

Finally, add the ports exposed above to the Ingress NGINX Controller Service:
```yaml
spec:
  ports:
  - nodePort: 31100
    port: 53
    name: kube-dns
```
Now, create a file called `ingress-nginx-udp-service-patch.yaml` with the YAML above.

Apply:
```bash
kubectl patch service ingress-nginx-udp-controller \ 
--patch "$(cat ingress-nginx-udp-service-patch.yaml)" \
--namespace ingress-nginx 
```

If everything went right, you should see something like the following:
```yaml
TODO
```

And should be able to test it with `telnet`, for example:
```bash
$ telnet 34.89.108.48 6379
Trying 34.89.108.48...
Connected to 34.89.108.48.
Escape character is '^]'.
```