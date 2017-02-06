# Deploying HAProxy Ingress Controller

Don't have a Kubernetes cluster? Single-node of [CoreOS Kubernetes](https://github.com/coreos/coreos-kubernetes/) is a good starting point.

Deploy a default backend used to serve `404 Not Found` pages:

    kubectl run ingress-default-backend \
      --image=gcr.io/google_containers/defaultbackend:1.0 \
      --port=8080 \
      --limits=cpu=10m,memory=20Mi \
      --expose

Check if the default backend is up and running:

    kubectl get pod
    NAME                                       READY     STATUS    RESTARTS   AGE
    ingress-default-backend-1110790216-gqr61   1/1       Running   0          10s

Deploy certificate and private key used to serve https on ingress that doesn't provide it's own certificate. For testing purposes a self signed certificate is ok:

    openssl req \
      -x509 -newkey rsa:2048 -nodes -days 365 \
      -keyout tls.key -out tls.crt -subj '/CN=localhost'
    kubectl create secret tls ingress-default-ssl --cert=tls.crt --key=tls.key
    rm -v tls.crt tls.key

Deploy HAProxy Ingress. Note that `hostNetwork: true` could be uncommented if your cluster has IPs that doesn't use ports 80, 443 and 1936.

    kubectl create -f haproxy-ingress.yaml

Check if the controller was successfully deployed:

    kubectl get pod -w
    NAME                                       READY     STATUS    RESTARTS   AGE
    haproxy-ingress-2556761959-tv20k           1/1       Running   0          12s
    ingress-default-backend-1110790216-gqr61   1/1       Running   0          3m
    ^C

Problem? Check logs and events of the POD:

    kubectl logs haproxy-ingress-2556761959-tv20k
    kubectl describe haproxy-ingress-2556761959-tv20k

Deploy some web application and it's ingress resource:

    kubectl run nginx --image=nginx:alpine --port=80 --expose
    kubectl create -f - <<EOF
      apiVersion: extensions/v1beta1
      kind: Ingress
      metadata:
        name: app
      spec:
        rules:
        - host: foo.bar
          http:
            paths:
            - path: /
              backend:
                serviceName: nginx
                servicePort: 80
    EOF

Exposing HAProxy Ingress depend on your Kubernetes environment. If `hostNetwork` was defined just use host's public IP, otherwise expose the controller as a `type=NodePort` service:

    kubectl expose deploy/haproxy-ingress --type=NodePort
    kubectl get svc/haproxy-ingress -oyaml

Look for `nodePort` field next to `port: 80`.

Change below `172.17.4.99` to the host's IP and `30876` to the `nodePort`, or remove `:30876` if using `hostNetwork`:

    curl -i 172.17.4.99:30876
    HTTP/1.1 404 Not Found
    Date: Mon, 05 Feb 2017 22:59:36 GMT
    Content-Length: 21
    Content-Type: text/plain; charset=utf-8

    default backend - 404

Using default backend because host was not found.

Now try to send a header:

    curl -i 172.17.4.99:30876 -H 'Host: foo.bar'
    HTTP/1.1 200 OK
    Server: nginx/1.11.9
    Date: Mon, 05 Feb 2017 23:00:33 GMT
    Content-Type: text/html
    Content-Length: 612
    Last-Modified: Tue, 24 Jan 2017 18:53:46 GMT
    ETag: "5887a2ba-264"
    Accept-Ranges: bytes

    <!DOCTYPE html>
    <html>
    <head>
    <title>Welcome to nginx!</title>
    ...

Not what you were looking for? Have a look at controller's logs:

    kubectl get pod
    NAME                                       READY     STATUS    RESTARTS   AGE
    haproxy-ingress-2556761959-tv20k           1/1       Running   0          9m
    ...

    kubectl logs haproxy-ingress-2556761959-tv20k | less -S
