# Nginx ingress controller using Proxy Protocol

For using the Proxy Protocol in a load balancing solution, both the load balancer and its backend need to enable Proxy Protocol.

To enable it for NGINX you have to setup a [configmap](nginx-configmap.yaml) option.

## HAProxy

This HAProxy snippet would forward HTTP(S) traffic to a two worker kubernetes cluster, with NGINX running on the node ports, like defined in this example's [service](nginx-svc.yaml).


```
listen kube-nginx-http
        bind :::80 v6only
        bind 0.0.0.0:80
        mode tcp
        option tcplog
        balance leastconn
        server node1 <node-ip1>:32080 check-send-proxy inter 10s send-proxy
        server node2 <node-ip2>:32080 check-send-proxy inter 10s send-proxy

listen kube-nginx-https
        bind :::443 v6only
        bind 0.0.0.0:443
        mode tcp
        option tcplog
        balance leastconn
        server node1 <node-ip1>:32443 check-send-proxy inter 10s send-proxy
        server node2 <node-ip2>:32443 check-send-proxy inter 10s send-proxy
```

## ELBs in AWS

See this [documentation](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html) how to enable Proxy Protocol in ELBs
