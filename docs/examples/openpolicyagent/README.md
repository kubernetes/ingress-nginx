# OpenPolicyAgent and pathType enforcing

Ingress API allows users to specify different [pathType](https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types) 
on Ingress object. 

While pathType `Exact` and `Prefix` should allow only a small set of characters, pathType `ImplementationSpecific`
allows any characters, as it may contain regexes, variables and other features that may be specific of the Ingress 
Controller being used.

This means that the Ingress Admins (the persona who deployed the Ingress Controller) should trust the users 
allowed to use `pathType: ImplementationSpecific`, as this may allow arbitrary configuration, and this 
configuration may end on the proxy (aka Nginx) configuration.

## Example
The example in this repo uses [Gatekeeper](https://open-policy-agent.github.io/gatekeeper/website/) to block the usage of `pathType: ImplementationSpecific`, 
allowing just a specific list of namespaces to use it.

It is recommended that the admin modifies this rules to enforce a specific set of characters when the usage of ImplementationSpecific
is allowed, or in ways that best suits their needs.

First, the `ConstraintTemplate` from [template.yaml](template.yaml) will define a rule that validates if the Ingress object 
is being created on an excempted namespace, and case not, will validate its pathType.

Then, the rule `K8sBlockIngressPathType` contained in [rule.yaml](rule.yaml) will define the parameters: what kind of 
object should be verified (Ingress), what are the excempted namespaces, and what kinds of pathType are blocked.
