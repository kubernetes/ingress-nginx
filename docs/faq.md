
# FAQ

## How can I easily install multiple instances of the ingress-nginx controller in the same cluster?

You can install them in different namespaces.

- Create a new namespace

  ```
  kubectl create namespace ingress-nginx-2
  ```

- Use Helm to install the additional instance of the ingress controller
- Ensure you have Helm working (refer to the [Helm documentation](https://helm.sh/docs/))
- We have to assume that you have the helm repo for the ingress-nginx controller already added to your Helm config.
  But, if you have not added the helm repo then you can do this to add the repo to your helm config;

  ```
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
  ```

- Make sure you have updated the helm repo data;

  ```
  helm repo update
  ```

- Now, install an additional instance of the ingress-nginx controller like this:

  ```
  helm install ingress-nginx-2 ingress-nginx/ingress-nginx  \
  --namespace ingress-nginx-2 \
  --set controller.ingressClassResource.name=nginx-two \
  --set controller.ingressClass=nginx-two \
  --set controller.ingressClassResource.controllerValue="example.com/ingress-nginx-2" \
  --set controller.ingressClassResource.enabled=true \
  --set controller.ingressClassByName=true
  ```

If you need to install yet another instance, then repeat the procedure to create a new namespace,
change the values such as names & namespaces (for example from "-2" to "-3"), or anything else that meets your needs.

Note that `controller.ingressClassResource.name` and `controller.ingressClass` have to be set correctly.
The first is to create the IngressClass object and the other is to modify the deployment of the actual ingress controller pod.

### I can't use multiple namespaces, what should I do?

If you need to install all instances in the same namespace, then you need to specify a different **election id**, like this:

```
helm install ingress-nginx-2 ingress-nginx/ingress-nginx  \
--namespace kube-system \
--set controller.electionID=nginx-two-leader \
--set controller.ingressClassResource.name=nginx-two \
--set controller.ingressClass=nginx-two \
--set controller.ingressClassResource.controllerValue="example.com/ingress-nginx-2" \
--set controller.ingressClassResource.enabled=true \
--set controller.ingressClassByName=true
```

## Retaining Client IPAddress

Please read [Retain Client IPAddress Guide here](./user-guide/retaining-client-ipaddress.md).

## Kubernetes v1.22 Migration

If you are using Ingress objects in your cluster (running Kubernetes older than
version 1.22), and you plan to upgrade your Kubernetes version to K8S 1.22 or
above, then please read [the migration guide here](./user-guide/k8s-122-migration.md).

## Validation Of **`path`**

- For improving security and also following desired standards on Kubernetes API
spec, the next release, scheduled for v1.8.0, will include a new & optional
feature of validating the value for the key `ingress.spec.rules.http.paths.path`.

- This behavior will be disabled by default on the 1.8.0 release and enabled by
default on the next breaking change release, set for 2.0.0.

- When "`ingress.spec.rules.http.pathType=Exact`" or "`pathType=Prefix`", this
validation will limit the characters accepted on the field "`ingress.spec.rules.http.paths.path`",
to "`alphanumeric characters`", and  `"/," "_," "-."` Also, in this case,
the path should start with `"/."`

- When the ingress resource path contains other characters (like on rewrite
configurations), the pathType value should be "`ImplementationSpecific`".

- API Spec on pathType is documented [here](https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types)

- When this option is enabled, the validation will happen on the Admission
Webhook. So if any new ingress object contains characters other than
alphanumeric characters, and, `"/,","_","-"`, in the `path` field, but
is not using `pathType` value as `ImplementationSpecific`, then the ingress
object will be denied admission.

- The cluster admin should establish validation rules using mechanisms like
"`Open Policy Agent`", to validate that only authorized users can use
ImplementationSpecific pathType and that only the authorized characters can be
used. [The configmap value is here](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#strict-validate-path-type)

- A complete example of an Openpolicyagent gatekeeper rule is available [here](https://kubernetes.github.io/ingress-nginx/examples/openpolicyagent/)

- If you have any issues or concerns, please do one of the following:
  - Open a GitHub issue
  - Comment in our Dev Slack Channel
  - Open a thread in our Google Group <ingress-nginx-dev@kubernetes.io>

## Why is chunking not working since controller v1.10 ?

- If your code is setting the HTTP header `"Transfer-Encoding: chunked"` and
the controller log messages show an error about duplicate header, it is
because of this change <http://hg.nginx.org/nginx/rev/2bf7792c262e>

- More details are available in this issue <https://github.com/kubernetes/ingress-nginx/issues/11162>
