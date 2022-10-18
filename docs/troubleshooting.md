<!--
-----------------NOTICE------------------------
This file is referenced in code as
https://github.com/kubernetes/ingress-nginx/blob/main/docs/troubleshooting.md
Do not move it without providing redirects.
-----------------------------------------------
-->

# Troubleshooting

## Ingress-Controller Logs and Events

There are many ways to troubleshoot the ingress-controller. The following are basic troubleshooting
methods to obtain more information.

### Check the Ingress Resource Events

```console
$ kubectl get ing -n <namespace-of-ingress-resource>
NAME           HOSTS      ADDRESS     PORTS     AGE
cafe-ingress   cafe.com   10.0.2.15   80        25s

$ kubectl describe ing <ingress-resource-name> -n <namespace-of-ingress-resource>
Name:             cafe-ingress
Namespace:        default
Address:          10.0.2.15
Default backend:  default-http-backend:80 (172.17.0.5:8080)
Rules:
  Host      Path  Backends
  ----      ----  --------
  cafe.com
            /tea      tea-svc:80 (<none>)
            /coffee   coffee-svc:80 (<none>)
Annotations:
  kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"annotations":{},"name":"cafe-ingress","namespace":"default","selfLink":"/apis/networking/v1/namespaces/default/ingresses/cafe-ingress"},"spec":{"rules":[{"host":"cafe.com","http":{"paths":[{"backend":{"serviceName":"tea-svc","servicePort":80},"path":"/tea"},{"backend":{"serviceName":"coffee-svc","servicePort":80},"path":"/coffee"}]}}]},"status":{"loadBalancer":{"ingress":[{"ip":"169.48.142.110"}]}}}

Events:
  Type    Reason  Age   From                      Message
  ----    ------  ----  ----                      -------
  Normal  CREATE  1m    ingress-nginx-controller  Ingress default/cafe-ingress
  Normal  UPDATE  58s   ingress-nginx-controller  Ingress default/cafe-ingress
```

### Check the Ingress Controller Logs

```console
$ kubectl get pods -n <namespace-of-ingress-controller>
NAME                                        READY     STATUS    RESTARTS   AGE
ingress-nginx-controller-67956bf89d-fv58j   1/1       Running   0          1m

$ kubectl logs -n <namespace> ingress-nginx-controller-67956bf89d-fv58j
-------------------------------------------------------------------------------
NGINX Ingress controller
  Release:    0.14.0
  Build:      git-734361d
  Repository: https://github.com/kubernetes/ingress-nginx
-------------------------------------------------------------------------------
....
```

### Check the Nginx Configuration

```console
$ kubectl get pods -n <namespace-of-ingress-controller>
NAME                                        READY     STATUS    RESTARTS   AGE
ingress-nginx-controller-67956bf89d-fv58j   1/1       Running   0          1m

$ kubectl exec -it -n <namespace-of-ingress-controller> ingress-nginx-controller-67956bf89d-fv58j -- cat /etc/nginx/nginx.conf
daemon off;
worker_processes 2;
pid /run/nginx.pid;
worker_rlimit_nofile 523264;
worker_shutdown_timeout 240s;
events {
	multi_accept        on;
	worker_connections  16384;
	use                 epoll;
}
http {
....
```

### Check if used Services Exist

```console
$ kubectl get svc --all-namespaces
NAMESPACE     NAME                   TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)         AGE
default       coffee-svc             ClusterIP   10.106.154.35    <none>        80/TCP          18m
default       kubernetes             ClusterIP   10.96.0.1        <none>        443/TCP         30m
default       tea-svc                ClusterIP   10.104.172.12    <none>        80/TCP          18m
kube-system   default-http-backend   NodePort    10.108.189.236   <none>        80:30001/TCP    30m
kube-system   kube-dns               ClusterIP   10.96.0.10       <none>        53/UDP,53/TCP   30m
kube-system   kubernetes-dashboard   NodePort    10.103.128.17    <none>        80:30000/TCP    30m
```

## Debug Logging

Using the flag `--v=XX` it is possible to increase the level of logging. This is performed by editing
the deployment.

```console
$ kubectl get deploy -n <namespace-of-ingress-controller>
NAME                       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
default-http-backend       1         1         1            1           35m
ingress-nginx-controller   1         1         1            1           35m

$ kubectl edit deploy -n <namespace-of-ingress-controller> ingress-nginx-controller
# Add --v=X to "- args", where X is an integer
```

- `--v=2` shows details using `diff` about the changes in the configuration in nginx
- `--v=3` shows details about the service, Ingress rule, endpoint changes and it dumps the nginx configuration in JSON format
- `--v=5` configures NGINX in [debug mode](https://nginx.org/en/docs/debugging_log.html)

## Authentication to the Kubernetes API Server

A number of components are involved in the authentication process and the first step is to narrow
down the source of the problem, namely whether it is a problem with service authentication or
with the kubeconfig file.

Both authentications must work:

```
+-------------+   service          +------------+
|             |   authentication   |            |
+  apiserver  +<-------------------+  ingress   |
|             |                    | controller |
+-------------+                    +------------+
```

**Service authentication**

The Ingress controller needs information from apiserver. Therefore, authentication is required, which can be achieved in a couple of ways:

* _Service Account:_ This is recommended, because nothing has to be configured. The Ingress controller will use information provided by the system to communicate with the API server. See 'Service Account' section for details.

* _Kubeconfig file:_ In some Kubernetes environments service accounts are not available. In this case a manual configuration is required. The Ingress controller binary can be started with the `--kubeconfig` flag. The value of the flag is a path to a file specifying how to connect to the API server. Using the `--kubeconfig` does not requires the flag `--apiserver-host`.
   The format of the file is identical to `~/.kube/config` which is used by kubectl to connect to the API server. See 'kubeconfig' section for details.

* _Using the flag `--apiserver-host`:_ Using this flag `--apiserver-host=http://localhost:8080` it is possible to specify an unsecured API server or reach a remote kubernetes cluster using [kubectl proxy](https://kubernetes.io/docs/user-guide/kubectl/kubectl_proxy/).
   Please do not use this approach in production.

In the diagram below you can see the full authentication flow with all options, starting with the browser
on the lower left hand side.

```
Kubernetes                                                  Workstation
+---------------------------------------------------+     +------------------+
|                                                   |     |                  |
|  +-----------+   apiserver        +------------+  |     |  +------------+  |
|  |           |   proxy            |            |  |     |  |            |  |
|  | apiserver |                    |  ingress   |  |     |  |  ingress   |  |
|  |           |                    | controller |  |     |  | controller |  |
|  |           |                    |            |  |     |  |            |  |
|  |           |                    |            |  |     |  |            |  |
|  |           |  service account/  |            |  |     |  |            |  |
|  |           |  kubeconfig        |            |  |     |  |            |  |
|  |           +<-------------------+            |  |     |  |            |  |
|  |           |                    |            |  |     |  |            |  |
|  +------+----+      kubeconfig    +------+-----+  |     |  +------+-----+  |
|         |<--------------------------------------------------------|        |
|                                                   |     |                  |
+---------------------------------------------------+     +------------------+
```

### Service Account

If using a service account to connect to the API server, the ingress-controller expects the file
`/var/run/secrets/kubernetes.io/serviceaccount/token` to be present. It provides a secret
token that is required to authenticate with the API server.

Verify with the following commands:

```console
# start a container that contains curl
$ kubectl run -it --rm test --image=curlimages/curl --restart=Never -- /bin/sh

# check if secret exists
/ $ ls /var/run/secrets/kubernetes.io/serviceaccount/
ca.crt     namespace  token
/ $

# check base connectivity from cluster inside
/ $ curl -k https://kubernetes.default.svc.cluster.local
{
  "kind": "Status",
  "apiVersion": "v1",
  "metadata": {

  },
  "status": "Failure",
  "message": "forbidden: User \"system:anonymous\" cannot get path \"/\"",
  "reason": "Forbidden",
  "details": {

  },
  "code": 403
}/ $

# connect using tokens
}/ $ curl --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt -H  "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" https://kubernetes.default.svc.cluster.local
&& echo
{
  "paths": [
    "/api",
    "/api/v1",
    "/apis",
    "/apis/",
    ... TRUNCATED
    "/readyz/shutdown",
    "/version"
  ]
}
/ $

# when you type `exit` or `^D` the test pod will be deleted.
```

If it is not working, there are two possible reasons:

1. The contents of the tokens are invalid. Find the secret name with `kubectl get secrets | grep service-account` and
   delete it with `kubectl delete secret <name>`. It will automatically be recreated.

2. You have a non-standard Kubernetes installation and the file containing the token may not be present.
   The API server will mount a volume containing this file, but only if the API server is configured to use
   the ServiceAccount admission controller.
   If you experience this error, verify that your API server is using the ServiceAccount admission controller.
   If you are configuring the API server by hand, you can set this with the `--admission-control` parameter.
   > Note that you should use other admission controllers as well. Before configuring this option, you should read about admission controllers.

More information:

- [User Guide: Service Accounts](http://kubernetes.io/docs/user-guide/service-accounts/)
- [Cluster Administrator Guide: Managing Service Accounts](http://kubernetes.io/docs/admin/service-accounts-admin/)

## Kube-Config

If you want to use a kubeconfig file for authentication, follow the [deploy procedure](deploy/index.md) and
add the flag `--kubeconfig=/etc/kubernetes/kubeconfig.yaml` to the args section of the deployment.

## Using GDB with Nginx

[Gdb](https://www.gnu.org/software/gdb/) can be used to with nginx to perform a configuration
dump. This allows us to see which configuration is being used, as well as older configurations.

Note: The below is based on the nginx [documentation](https://docs.nginx.com/nginx/admin-guide/monitoring/debugging/#dumping-nginx-configuration-from-a-running-process).

1. SSH into the worker

    ```console
    $ ssh user@workerIP
    ```

2. Obtain the Docker Container Running nginx

    ```console
    $ docker ps | grep ingress-nginx-controller
    CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
    d9e1d243156a        registry.k8s.io/ingress-nginx/controller   "/usr/bin/dumb-init …"   19 minutes ago      Up 19 minutes                                                                            k8s_ingress-nginx-controller_ingress-nginx-controller-67956bf89d-mqxzt_kube-system_079f31ec-aa37-11e8-ad39-080027a227db_0
    ```

3. Exec into the container

    ```console
    $ docker exec -it --user=0 --privileged d9e1d243156a bash
    ```

4. Make sure nginx is running in `--with-debug`

    ```console
    $ nginx -V 2>&1 | grep -- '--with-debug'
    ```

5. Get list of processes running on container

    ```console
    $ ps -ef
    UID        PID  PPID  C STIME TTY          TIME CMD
    root         1     0  0 20:23 ?        00:00:00 /usr/bin/dumb-init /nginx-ingres
    root         5     1  0 20:23 ?        00:00:05 /ingress-nginx-controller --defa
    root        21     5  0 20:23 ?        00:00:00 nginx: master process /usr/sbin/
    nobody     106    21  0 20:23 ?        00:00:00 nginx: worker process
    nobody     107    21  0 20:23 ?        00:00:00 nginx: worker process
    root       172     0  0 20:43 pts/0    00:00:00 bash
    ```

6. Attach gdb to the nginx master process

    ```console
    $ gdb -p 21
    ....
    Attaching to process 21
    Reading symbols from /usr/sbin/nginx...done.
    ....
    (gdb)
    ```

7. Copy and paste the following:

    ```console
    set $cd = ngx_cycle->config_dump
    set $nelts = $cd.nelts
    set $elts = (ngx_conf_dump_t*)($cd.elts)
    while ($nelts-- > 0)
    set $name = $elts[$nelts]->name.data
    printf "Dumping %s to nginx_conf.txt\n", $name
    append memory nginx_conf.txt \
            $elts[$nelts]->buffer.start $elts[$nelts]->buffer.end
    end
    ```

8. Quit GDB by pressing CTRL+D

9. Open nginx_conf.txt

    ```console
    cat nginx_conf.txt
    ```
    
## Image related issues faced on Nginx 4.2.5 or other versions (Helm chart versions) 

1. Incase you face below error while installing Nginx using helm chart (either by helm commands or helm_release terraform provider ) 
```
Warning  Failed     5m5s (x4 over 6m34s)   kubelet            Failed to pull image "registry.k8s.io/ingress-nginx/kube-webhook-certgen:v1.3.0@sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47": rpc error: code = Unknown desc = failed to pull and unpack image "registry.k8s.io/ingress-nginx/kube-webhook-certgen@sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47": failed to resolve reference "registry.k8s.io/ingress-nginx/kube-webhook-certgen@sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47": failed to do request: Head "https://eu.gcr.io/v2/k8s-artifacts-prod/ingress-nginx/kube-webhook-certgen/manifests/sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47": EOF
```
   Then please follow the below steps.

2. During troubleshooting you can also execute the below commands to test the connectivities from you local machines and repositories  details

      a. curl registry.k8s.io/ingress-nginx/kube-webhook-certgen@sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47 > /dev/null
      ```
      (⎈ |myprompt)➜  ~ curl registry.k8s.io/ingress-nginx/kube-webhook-certgen@sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47 > /dev/null
                          % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                                          Dload  Upload   Total   Spent    Left  Speed
                          0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
       (⎈ |myprompt)➜  ~
      ```
      b. curl -I https://eu.gcr.io/v2/k8s-artifacts-prod/ingress-nginx/kube-webhook-certgen/manifests/sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47
      ```
      (⎈ |myprompt)➜  ~ curl -I https://eu.gcr.io/v2/k8s-artifacts-prod/ingress-nginx/kube-webhook-certgen/manifests/sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47
                                          HTTP/2 200
                                          docker-distribution-api-version: registry/2.0
                                          content-type: application/vnd.docker.distribution.manifest.list.v2+json
                                          docker-content-digest: sha256:549e71a6ca248c5abd51cdb73dbc3083df62cf92ed5e6147c780e30f7e007a47
                                          content-length: 1384
                                          date: Wed, 28 Sep 2022 16:46:28 GMT
                                          server: Docker Registry
                                          x-xss-protection: 0
                                          x-frame-options: SAMEORIGIN
                                          alt-svc: h3=":443"; ma=2592000,h3-29=":443"; ma=2592000,h3-Q050=":443"; ma=2592000,h3-Q046=":443"; ma=2592000,h3-Q043=":443"; ma=2592000,quic=":443"; ma=2592000; v="46,43"

        (⎈ |myprompt)➜  ~
      ```
   Redirection in the proxy is implemented to ensure the pulling of the images.

3. This is the solution recommended to whitelist the below image repositories : 
     ```
     *.appspot.com    
     *.k8s.io        
     *.pkg.dev
     *.gcr.io
     
     ```
     More details about the above repos : 
     a. *.k8s.io -> To ensure you can pull any images from registry.k8s.io
     b. *.gcr.io -> GCP services are used for image hosting. This is part of the domains suggested by GCP to allow and ensure users can pull images from their container registry services.
     c. *.appspot.com -> This a Google domain. part of the domain used for GCR.

## Unable to listen on port (80/443)
One possible reason for this error is lack of permission to bind to the port.  Ports 80, 443, and any other port < 1024 are Linux privileged ports which historically could only be bound by root.  The ingress-nginx-controller uses the CAP_NET_BIND_SERVICE [linux capability](https://man7.org/linux/man-pages/man7/capabilities.7.html) to allow binding these ports as a normal user (www-data / 101).  This involves two components:
1. In the image, the /nginx-ingress-controller file has the cap_net_bind_service capability added (e.g. via [setcap](https://man7.org/linux/man-pages/man8/setcap.8.html)) 
2. The NET_BIND_SERVICE capability is added to the container in the containerSecurityContext of the deployment.

If encountering this on one/some node(s) and not on others, try to purge and pull a fresh copy of the image to the affected node(s), in case there has been corruption of the underlying layers to lose the capability on the executable.

### Create a test pod
The /nginx-ingress-controller process exits/crashes when encountering this error, making it difficult to troubleshoot what is happening inside the container.  To get around this, start an equivalent container running "sleep 3600", and exec into it for further troubleshooting.  For example:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: ingress-nginx-sleep
  namespace: default
  labels:
    app: nginx
spec:
  containers:
    - name: nginx
      image: ##_CONTROLLER_IMAGE_##
      resources:
        requests:
          memory: "512Mi"
          cpu: "500m"
        limits:
          memory: "1Gi"
          cpu: "1"
      command: ["sleep"]
      args: ["3600"]
      ports:
      - containerPort: 80
        name: http
        protocol: TCP
      - containerPort: 443
        name: https
        protocol: TCP
      securityContext:
        allowPrivilegeEscalation: true
        capabilities:
          add:
          - NET_BIND_SERVICE
          drop:
          - ALL
        runAsUser: 101
  restartPolicy: Never
  nodeSelector:
    kubernetes.io/hostname: ##_NODE_NAME_##
  tolerations:
  - key: "node.kubernetes.io/unschedulable"
    operator: "Exists"
    effect: NoSchedule
```
* update the namespace if applicable/desired
* replace `##_NODE_NAME_##` with the problematic node (or remove nodeSelector section if problem is not confined to one node)
* replace `##_CONTROLLER_IMAGE_##` with the same image as in use by your ingress-nginx deployment
* confirm the securityContext section matches what is in place for ingress-nginx-controller pods in your cluster

Apply the YAML and open a shell into the pod.
Try to manually run the controller process:
```console
$ /nginx-ingress-controller
```
You should get the same error as from the ingress controller pod logs.

Confirm the capabilities are properly surfacing into the pod:
```console
$ grep CapBnd /proc/1/status
CapBnd: 0000000000000400
```
The above value has only net_bind_service enabled (per security context in YAML which adds that and drops all). If you get a different value, then you can decode it on another linux box (capsh not available in this container) like below, and then figure out why specified capabilities are not propagating into the pod/container.
```console
$ capsh --decode=0000000000000400
0x0000000000000400=cap_net_bind_service
```

## Create a test pod as root
(Note, this may be restricted by PodSecurityPolicy, PodSecurityAdmission/Standards, OPA Gatekeeper, etc. in which case you will need to do the appropriate workaround for testing, e.g. deploy in a new namespace without the restrictions.)
To test further you may want to install additional utilities, etc.  Modify the pod yaml by:
* changing runAsUser from 101 to 0
* removing the "drop..ALL" section from the capabilities.

Some things to try after shelling into this container:

Try running the controller as the www-data (101) user:
```console
$ chmod 4755 /nginx-ingress-controller
$ /nginx-ingress-controller
```
Examine the errors to see if there is still an issue listening on the port or if it passed that and moved on to other expected errors due to running out of context.

Install the libcap package and check capabilities on the file:
```console
$ apk add libcap
(1/1) Installing libcap (2.50-r0)
Executing busybox-1.33.1-r7.trigger
OK: 26 MiB in 41 packages
$ getcap /nginx-ingress-controller
/nginx-ingress-controller cap_net_bind_service=ep
```
(if missing, see above about purging image on the server and re-pulling)

Strace the executable to see what system calls are being executed when it fails:
```console
$ apk add strace
(1/1) Installing strace (5.12-r0)
Executing busybox-1.33.1-r7.trigger
OK: 28 MiB in 42 packages
$ strace /nginx-ingress-controller
execve("/nginx-ingress-controller", ["/nginx-ingress-controller"], 0x7ffeb9eb3240 /* 131 vars */) = 0
arch_prctl(ARCH_SET_FS, 0x29ea690)      = 0
...
```
