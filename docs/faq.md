
# FAQ

## Retaining Client IPAddress

Please read [Retain Client IPAddress Guide here](./user-guide/retaining-client-ipaddress.md).

## Kubernetes v1.22 Migration

If you are using Ingress objects in your cluster (running Kubernetes older than v1.22), and you plan to upgrade your Kubernetes version to K8S 1.22 or above, then please read [the migration guide here](./user-guide/k8s-122-migration.md).

## Validation Of __`path`__

- For improving security and also following desired standards on Kubernetes API spec, the next release, scheduled for v1.8.0, will include a new & optional feature of validating the value for the key `ingress.spec.rules.http.paths.path` .

- This behavior will be disabled by default on the 1.8.0 release and enabled by default on the next breaking change release, set for 2.0.0.

- When "`ingress.spec.rules.http.pathType=Exact`" or "`pathType=Prefix`", this validation will limit the characters accepted on the field "`ingress.spec.rules.http.paths.path`",  to "`alphanumeric characters`", and  `"/," "_," "-."` Also, in this case, the path should start with `"/."`

- When the ingress resource path contains other characters (like on rewrite configurations), the pathType value should be "`ImplementationSpecific`".

- API Spec on pathType is documented [here](https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types)

- When this option is enabled, the validation will happen on the Admission Webhook. So if any new ingress object contains characters other than  "`alphanumeric characters`", and  `"/," "_," "-."` , in the `path` field, but is not using `pathType` value as `ImplementationSpecific`, then the ingress object will be denied admission.

- The cluster admin should establish validation rules using mechanisms like "`Open Policy Agent`", to validate that only authorized users can use ImplementationSpecific pathType and that only the authorized characters can be used. [The configmap value is here](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#strict-validate-path-type)

- A complete example of an Openpolicyagent gatekeeper rule is available [here](https://kubernetes.github.io/ingress-nginx/examples/openpolicyagent/)

- If you have any issues or concerns, please do one of the following: 
  - Open a GitHub issue 
  - Comment in our Dev Slack Channel
  - Open a thread in our Google Group ingress-nginx-dev@kubernetes.io
