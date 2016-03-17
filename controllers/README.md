# Ingress Controllers

Configuring a webserver or loadbalancer is harder than it should be. Most webserver configuration files are very similar. There are some applications that have weird little quirks that tend to throw a wrench in things, but for the most part you can apply the same logic to them and achieve a desired result. The Ingress resource embodies this idea, and an Ingress controller is meant to handle all the quirks associated with a specific "class" of Ingress (be it a single instance of a loadbalancer, or a more complicated setup of frontends that provide GSLB, DDoS protection etc).

## What is an Ingress Controller?

An Ingress Controller is a daemon, deployed as a Kubernetes Pod, that watches the ApiServer's `/ingresses` endpoint for updates to the [Ingress resource](https://github.com/kubernetes/kubernetes/blob/master/docs/user-guide/ingress.md). Its job is to satisfy requests for ingress.

## Writing an Ingress Controller

Writing an Ingress controller is simple. By way of example, the [nginx controller] (nginx-alpha) does the following:
* Poll until apiserver reports a new Ingress
* Write the nginx config file based on a [go text/template](https://golang.org/pkg/text/template/)
* Reload nginx

Pay attention to how it denormalizes the Kubernetes Ingress object into an nginx config:
```go
const (
	nginxConf = `
events {
  worker_connections 1024;
}
http {
{{range $ing := .Items}}
{{range $rule := $ing.Spec.Rules}}
  server {
    listen 80;
    server_name {{$rule.Host}};
{{ range $path := $rule.HTTP.Paths }}
    location {{$path.Path}} {
      proxy_set_header Host $host;
      proxy_pass http://{{$path.Backend.ServiceName}}.{{$ing.Namespace}}.svc.cluster.local:{{$path.Backend.ServicePort}};
    }{{end}}
  }{{end}}{{end}}
}`
)
```

You can take a similar approach to denormalize the Ingress to a [haproxy config](https://github.com/kubernetes/contrib/blob/master/service-loadbalancer/template.cfg) or use it to configure a cloud loadbalancer such as a [GCE L7](https://github.com/kubernetes/contrib/blob/master/ingress/controllers/gce/README.md).

And here is the Ingress controller's control loop:

```go
for {
	rateLimiter.Accept()
	ingresses, err := ingClient.List(labels.Everything(), fields.Everything())
	if err != nil || reflect.DeepEqual(ingresses.Items, known.Items) {
	    continue
	}
	if w, err := os.Create("/etc/nginx/nginx.conf"); err != nil {
		log.Fatalf("Failed to open %v: %v", nginxConf, err)
	} else if err := tmpl.Execute(w, ingresses); err != nil {
		log.Fatalf("Failed to write template %v", err)
	}
	shellOut("nginx -s reload")
}
```

All this is doing is:
* List Ingresses, optionally you can watch for changes (see [GCE Ingress controller](https://github.com/kubernetes/contrib/blob/master/ingress/controllers/gce/controller.go) for an example)
* Executes the template and writes results to `/etc/nginx/nginx.conf`
* Reloads nginx

You can deploy this controller to a Kubernetes cluster by [creating an RC](nginx-alpha/rc.yaml). After doing so, if you were to create an Ingress such as:
```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: /foo
        backend:
          serviceName: fooSvc
          servicePort: 80
  - host: bar.baz.com
    http:
      paths:
      - path: /bar
        backend:
          serviceName: barSvc
          servicePort: 80
```

Where `fooSvc` and `barSvc` are 2 services running in your Kubernetes cluster. The controller would satisfy the Ingress by writing a configuration file to /etc/nginx/nginx.conf:
```nginx
events {
  worker_connections 1024;
}
http {
  server {
    listen 80;
    server_name foo.bar.com;

    location /foo {
      proxy_pass http://fooSvc;
    }
  }
  server {
    listen 80;
    server_name bar.baz.com;

    location /bar {
      proxy_pass http://barSvc;
    }
  }
}
```

And you can reach the `/foo` and `/bar` endpoints on the publicIP of the VM the nginx-ingress pod landed on.
```
$ kubectl get pods -o wide
NAME                  READY     STATUS    RESTARTS   AGE       NODE
nginx-ingress-tk7dl   1/1       Running   0          3m        e2e-test-beeps-minion-15p3

$ kubectl get nodes e2e-test-beeps-minion-15p3 -o yaml | grep -i externalip -B 1
  - address: 104.197.203.179
    type: ExternalIP

$ curl --resolve foo.bar.com:80:104.197.203.179 foo.bar.com/foo
```

## Future work

This section can also bear the title "why anyone would want to write an Ingress controller instead of directly configuring Services". There is more to Ingress than webserver configuration. *Real* HA usually involves the configuration of gateways and packet forwarding devices, which most cloud providers allow you to do through an API. See the GCE Loadbalancer Controller, which is deployed as a [cluster addon](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/cluster-loadbalancing/glbc) in GCE and GKE clusters for more advanced Ingress configuration examples. Post 1.2 the Ingress resource will support at least the following:
* More TLS options (SNI, re-encrypt etc)
* L4 and L7 loadbalancing (it currently only supports HTTP rules)
* Ingress Rules that are not limited to a simple path regex (eg: redirect rules, session persistence)

And is expected to be the way one configures a "frontends" that handle user traffic for a Kubernetes cluster.


