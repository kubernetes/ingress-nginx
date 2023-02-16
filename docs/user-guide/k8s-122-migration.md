# FAQ - Migration to Kubernetes 1.22 and apiVersion `networking.k8s.io/v1`

If you are using Ingress objects in your cluster (running Kubernetes older than v1.22),
and you plan to upgrade to Kubernetes v1.22, this page is relevant to you.

- Please read this [official blog on deprecated Ingress API versions](https://kubernetes.io/blog/2021/07/26/update-with-ingress-nginx/)
- Please read this [official documentation on the IngressClass object](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)

## What is an IngressClass and why is it important for users of ingress-nginx controller now?

IngressClass is a Kubernetes resource. See the description below.
It's important because until now, a default install of the ingress-nginx controller did not require a IngressClass object.
From version 1.0.0 of the ingress-nginx controller, an IngressClass object is required.

On clusters with more than one instance of the ingress-nginx controller, all instances of the controllers must be aware of which Ingress objects they serve.
The `ingressClassName` field of an Ingress is the way to let the controller know about that.

```console
kubectl explain ingressclass
```

```
KIND:     IngressClass
VERSION:  networking.k8s.io/v1
DESCRIPTION:
     IngressClass represents the class of the Ingress, referenced by the Ingress
     Spec. The `ingressclass.kubernetes.io/is-default-class` annotation can be
     used to indicate that an IngressClass should be considered default. When a
     single IngressClass resource has this annotation set to true, new Ingress
     resources without a class specified will be assigned this default class.
FIELDS:
   apiVersion   <string>
     APIVersion defines the versioned schema of this representation of an
     object. Servers should convert recognized schemas to the latest internal
     value, and may reject unrecognized values. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
   kind <string>
     Kind is a string value representing the REST resource this object
     represents. Servers may infer this from the endpoint the client submits
     requests to. Cannot be updated. In CamelCase. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
   metadata     <Object>
     Standard object's metadata. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
   spec <Object>
     Spec is the desired state of the IngressClass. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status`
```

## What has caused this change in behavior?

There are 2 primary reasons.

### Reason 1

Until K8s version 1.21, it was possible to create an Ingress resource using deprecated versions of the Ingress API, such as:

- `extensions/v1beta1`
- `networking.k8s.io/v1beta1`
  You would get a message about deprecation, but the Ingress resource would get created.

From K8s version 1.22 onwards, you can **only** access the Ingress API via the stable, `networking.k8s.io/v1` API.
The reason is explained in the [official blog on deprecated ingress API versions](https://kubernetes.io/blog/2021/07/26/update-with-ingress-nginx/).

### Reason #2

If you are already using the ingress-nginx controller and then upgrade to Kubernetes 1.22,
there are several scenarios where your existing Ingress objects will not work how you expect.

Read this FAQ to check which scenario matches your use case.

## What is the `ingressClassName` field?

`ingressClassName` is a field in the spec of an Ingress object.

```shell
kubectl explain ingress.spec.ingressClassName
```

```console
KIND:     Ingress
VERSION:  networking.k8s.io/v1
FIELD:    ingressClassName <string>
DESCRIPTION:
     IngressClassName is the name of the IngressClass cluster resource. The
     associated IngressClass defines which controller will implement the
     resource. This replaces the deprecated `kubernetes.io/ingress.class`
     annotation. For backwards compatibility, when that annotation is set, it
     must be given precedence over this field. The controller may emit a warning
     if the field and annotation have different values. Implementations of this
     API should ignore Ingresses without a class specified. An IngressClass
     resource may be marked as default, which can be used to set a default value
     for this field. For more information, refer to the IngressClass
     documentation.
```

The `.spec.ingressClassName` behavior has precedence over the deprecated `kubernetes.io/ingress.class` annotation.

## I have only one ingress controller in my cluster. What should I do?

If a single instance of the ingress-nginx controller is the sole Ingress controller running in your cluster,
you should add the annotation "ingressclass.kubernetes.io/is-default-class" in your IngressClass,
so any new Ingress objects will have this one as default IngressClass.

When using Helm, you can enable this annotation by setting `.controller.ingressClassResource.default: true` in your Helm chart installation's values file.

If you have any old Ingress objects remaining without an IngressClass set, you can do one or more of the following to make the ingress-nginx controller aware of the old objects:

- You can manually set the [`.spec.ingressClassName`](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/ingress-v1/#IngressSpec) field in the manifest of your own Ingress resources.
- You can re-create them after setting the `ingressclass.kubernetes.io/is-default-class` annotation to `true` on the IngressClass
- Alternatively you can make the ingress-nginx controller watch Ingress objects without the ingressClassName field set by starting your ingress-nginx with the flag [--watch-ingress-without-class=true](#what-is-the-flag-watch-ingress-without-class).
  When using Helm, you can configure your Helm chart installation's values file with `.controller.watchIngressWithoutClass: true`.

We recommend that you create the IngressClass as shown below:

```
---
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  labels:
    app.kubernetes.io/component: controller
  name: nginx
  annotations:
    ingressclass.kubernetes.io/is-default-class: "true"
spec:
  controller: k8s.io/ingress-nginx
```

and add the value `spec.ingressClassName=nginx` in your Ingress objects.

## I have many ingress objects in my cluster. What should I do?

If you have a lot of ingress objects without ingressClass configuration,
you can run the ingress controller with the flag `--watch-ingress-without-class=true`.

### What is the flag `--watch-ingress-without-class`?

It's a flag that is passed, as an argument, to the `nginx-ingress-controller` executable.
In the configuration, it looks like this:

```yaml
# ...
args:
  - /nginx-ingress-controller
  - --watch-ingress-without-class=true
  - --controller-class=k8s.io/ingress-nginx
  # ...
# ...
```

## I have more than one controller in my cluster, and I'm already using the annotation

No problem. This should still keep working, but we highly recommend you to test!
Even though `kubernetes.io/ingress.class` is deprecated, the ingress-nginx controller still understands that annotation.
If you want to follow good practice, you should consider migrating to use IngressClass and `.spec.ingressClassName`.

## I have more than one controller running in my cluster, and I want to use the new API

In this scenario, you need to create multiple IngressClasses (see the example above).

Be aware that IngressClass works in a very specific way: you will need to change the `.spec.controller` value in your IngressClass and configure the controller to expect the exact same value.

Let's see an example, supposing that you have three IngressClasses:

- IngressClass `ingress-nginx-one`, with `.spec.controller` equal to `example.com/ingress-nginx1`
- IngressClass `ingress-nginx-two`, with `.spec.controller` equal to `example.com/ingress-nginx2`
- IngressClass `ingress-nginx-three`, with `.spec.controller` equal to `example.com/ingress-nginx1`

For private use, you can also use a controller name that doesn't contain a `/`, e.g. `ingress-nginx1`.

When deploying your ingress controllers, you will have to change the `--controller-class` field as follows:

- Ingress-Nginx A, configured to use controller class name `example.com/ingress-nginx1`
- Ingress-Nginx B, configured to use controller class name `example.com/ingress-nginx2`

When you create an Ingress object with its `ingressClassName` set to `ingress-nginx-two`,
only controllers looking for the `example.com/ingress-nginx2` controller class pay attention to the new object.

Given that Ingress-Nginx B is set up that way, it will serve that object, whereas Ingress-Nginx A ignores the new Ingress.

Bear in mind that if you start Ingress-Nginx B with the command line argument `--watch-ingress-without-class=true`, it will serve:

1. Ingresses without any `ingressClassName` set
2. Ingresses where the deprecated annotation (`kubernetes.io/ingress.class`) matches the value set in the command line argument `--ingress-class`
3. Ingresses that refer to any IngressClass that has the same `spec.controller` as configured in `--controller-class`
4. If you start Ingress-Nginx B with the command line argument `--watch-ingress-without-class=true` and you run Ingress-Nginx A with the command line argument `--watch-ingress-without-class=false` then this is a supported configuration.
   If you have two ingress-nginx controllers for the same cluster, both running with `--watch-ingress-without-class=true` then there is likely to be a conflict.

## Why am I seeing "ingress class annotation is not equal to the expected by Ingress Controller" in my controller logs?

It is highly likely that you will also see the name of the ingress resource in the same error message.
This error message has been observed on use the deprecated annotation (`kubernetes.io/ingress.class`) in an Ingress resource manifest.
It is recommended to use the `.spec.ingressClassName` field of the Ingress resource, to specify the name of the IngressClass of the Ingress you are defining.

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
