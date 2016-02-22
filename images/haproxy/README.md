
HAProxy 1.6 base image using alpine linux

What is HAProxy?
HAProxy is a free, very fast and reliable solution offering high availability, load balancing, and proxying for TCP and HTTP-based applications.

**How to use this image:**
This image does provides a default configuration file with no backend servers.

*Using docker*
```
$ docker run -v /some/haproxy.cfg:/etc/haproxy/haproxy.cfg:ro gcr.io/google_containers/haproxy:0.2
```

*Creating a pod*
```
$ kubectl create -f ./pod.yaml
```
