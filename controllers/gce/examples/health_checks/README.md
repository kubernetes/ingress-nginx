# Simple HTTP health check example

The GCE Ingress controller adopts the readiness probe from the matching endpoints, provided the readiness probe doesn't require special headers.

Create the following app:
```console
$ kubectl create -f health_check_app.yaml
replicationcontroller "echoheaders" created
You have exposed your service on an external port on all nodes in your
cluster.  If you want to expose this service to the external internet, you may
need to set up firewall rules for the service port(s) (tcp:31165) to serve traffic.

See http://releases.k8s.io/HEAD/docs/user-guide/services-firewalls.md for more details.
service "echoheadersx" created
You have exposed your service on an external port on all nodes in your
cluster.  If you want to expose this service to the external internet, you may
need to set up firewall rules for the service port(s) (tcp:31020) to serve traffic.

See http://releases.k8s.io/HEAD/docs/user-guide/services-firewalls.md for more details.
service "echoheadersy" created
ingress "echomap" created
```

You should soon find an Ingress that is backed by a GCE Loadbalancer.

```console
$ kubectl describe ing echomap
Name:			echomap
Namespace:		default
Address:		107.178.255.228
Default backend:	default-http-backend:80 (10.180.0.9:8080,10.240.0.2:8080)
Rules:
  Host		Path	Backends
  ----		----	--------
  foo.bar.com
    		/foo 	echoheadersx:80 (<none>)
  bar.baz.com
    		/bar 	echoheadersy:80 (<none>)
    		/foo 	echoheadersx:80 (<none>)
Annotations:
  target-proxy:		k8s-tp-default-echomap--a9d60e8176d933ee
  url-map:		k8s-um-default-echomap--a9d60e8176d933ee
  backends:		{"k8s-be-31020--a9d60e8176d933ee":"HEALTHY","k8s-be-31165--a9d60e8176d933ee":"HEALTHY","k8s-be-31686--a9d60e8176d933ee":"HEALTHY"}
  forwarding-rule:	k8s-fw-default-echomap--a9d60e8176d933ee
Events:
  FirstSeen	LastSeen	Count	From				SubobjectPath	Type		Reason	Message
  ---------	--------	-----	----				-------------	--------	------	-------
  17m		17m		1	{loadbalancer-controller }			Normal		ADD	default/echomap
  15m		15m		1	{loadbalancer-controller }			Normal		CREATE	ip: 107.178.255.228

$ curl 107.178.255.228/foo -H 'Host:foo.bar.com'
CLIENT VALUES:
client_address=10.240.0.5
command=GET
real path=/foo
query=nil
request_version=1.1
request_uri=http://foo.bar.com:8080/foo
...
```

You can confirm the health check endpoint point it's using one of 2 ways:
* Through the cloud console: compute > health checks > lookup your health check. It takes the form k8s-be-nodePort-hash, where nodePort in the example above is 31165 and 31020, as shown by the kubectl output.
* Through gcloud: Run `gcloud compute http-health-checks list`

## Limitations

A few points to note:
* The readiness probe must be exposed on the port matching the `servicePort` specified in the Ingress
* The readiness probe cannot have special requirements like headers
* The probe timeouts are translated to GCE health check timeouts
* You must create the pods backing the endpoints with the given readiness probe. This *will not* work if you update the replication controller with a different readiness probe.
