# Pod Security Policy (PSP)

In most clusters today, by default, all resources (e.g. Deployments and ReplicatSets)
have permissions to create pods.
Kubernetes however provides a more fine-grained authorization policy called 
[Pod Security Policy (PSP)](https://kubernetes.io/docs/concepts/policy/pod-security-policy/).

PSP allows the cluster owner to define the permission of each object, for example creating a pod.
If you have PSP enabled on the cluster, and you deploy ingress-nginx,
you will need to provide the Deployment with the permissions to create pods.

Before applying any objects, first apply the PSP permissions by running:
```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/docs/examples/psp/psp.yaml
```

Now that the pod security policy is applied, we can continue as usual by applying the
[mandatory.yaml](https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/mandatory.yaml)
according to the [Installation Guide](../../deploy/index.md). 

Note: PSP permissions must be granted before to the creation of the Deployment and the ReplicaSet.
If the Deployment or ReplicaSet already exist, they will receive the PSP permissions
only after deleting them and reapplying mandatory.yaml.
