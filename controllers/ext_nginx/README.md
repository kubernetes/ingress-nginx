# External Nginx Ingress Controller

This is a nginx ingress controller that will run outside Kubernetes but still update nginx configuration files.

It functions very much like https://github.com/kubernetes/ingress/tree/master/controllers/nginx but runs as a standalone process that just manages nginx.conf.

## Why?

The user case where you host your kubernetes cluster on-premise and don't want to have multiple layers of "load balancing" in front of your pods. You also really like nginx. 
Using this ingress controller on your edge load balacing cluster, the configuration is kept up too date and you can still use nginx build in functions for zero down time deployments of config and binaries that the normal "in-cluster" ingress controller don't support.

This ingress controller is NOT to be used if you don't understand exaktly what you are doing.

## Usage

You use it very much like the original nginx ingress controller. See: https://github.com/kubernetes/ingress/tree/master/controllers/nginx for full details.  Except you should run it as a system service. 

example:
`export POD_NAME=default`
`export POD_NAMESPACE=default`
`./nginx-ingress-controller --default-backend-service=default/$K8_DEFAULTBACKEND --apiserver-host=https://$K8_API --kubeconfig ~/.kube/config`


