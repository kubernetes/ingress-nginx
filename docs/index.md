# Welcome

This is the documentation for the NGINX Ingress Controller.

It is built around the [Kubernetes Ingress resource](http://kubernetes.io/docs/user-guide/ingress/), using a [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#understanding-configmaps-and-pods) to store the NGINX configuration.

Learn more about using Ingress on [k8s.io](http://kubernetes.io/docs/user-guide/ingress/).

## Getting Started

See [Deployment](./deploy/) for a whirlwind tour that will get you started.


# FAQ - Migration to apiVersion networking.k8s.io/v1

- Please read this [official blog on deprecated ingress api versions](https://kubernetes.io/blog/2021/07/26/update-with-ingress-nginx/) If you are using ingress objects in your pre K8s v1.22 cluster, and you upgrade to K8s v1.22, then this document may be relevant to you.

- Please read this [official documentation on the IngressClass object](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)

## What is an ingressClass and why is it important for users of Ingress-NGINX controller now ?

IngressClass is a Kubernetes resource. See the description below.
Its important because until now, a default install of the Ingress-NGINX controller did not require a ingressClass object. But from version 1.0.0 of the Ingress-NGINX Controller, a ingressclass object is required.

On clusters with more than one instance of the Ingress-NGINX controller, all instances of the controllers must be aware of which Ingress object they must serve. The ingressClass field of a ingress object is the way to let the controller know about that. 

```
_$ k explain ingressClass                                                           
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

## What has caused this change in behaviour ?

There are 2 reasons primarily.

(Reason #1) Until K8s version 1.21, it was was possible to create a ingress resource, with the "apiVersion:" field set to a value like ;
  - extensions/v1beta1
  - networking.k8s.io/v1beta1

    (You would get a message about deprecation but the ingress resource would get created.)

From K8s version 1.22 onwards, you can ONLY set the "apiVersion:" field of a ingress resource, to the value "networking.k8s.io/v1". The reason is [official blog on deprecated ingress api versions](https://kubernetes.io/blog/2021/07/26/update-with-ingress-nginx/).

(Reason #2) When you upgrade to K8s version v1.22, while you are already using the Ingress-NGINX controller, there are several scenarios where the old existing ingress objects will not work. Read this FAQ to check which scenario matches your use case.

## What is ingressClassName field ?

ingressClassName is a field in the specs of a ingress object.

```
% k explain ingress.spec.ingressClassName
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
 the spec.ingressClassName behavior has precedence over the annotation.



## I have only one instance of the Ingresss-NGINX controller in my cluster. What should I do ?

- If you have only one instance of the Ingress-NGINX controller running in your cluster, and you still want to use ingressclass, you should add the annotation "ingressclass.kubernetes.io/is-default-class" in your ingress class, so any new Ingress objects will have this one as default ingressClass.

In this case, you need to make your Controller aware of the objects. If you have several Ingress objects and they don't yet have the [ingressClassName](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#ingress-v1-networking-k8s-io) field, or the ingress annotation (`kubernetes.io/ingress.class`),  then you should start your ingress-controller with the flag [--watch-ingress-without-class=true](## What is the flag '--watch-ingress-without-class' ?) .

You can configure your helm chart installation's values file with `.controller.watchIngressWithoutClass: true`. 

We highly recommend that you create the ingressClass as shown below:
```
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
And add the value "spec.ingressClassName=nginx" in your Ingress objects

## I have multiple ingress objects in my cluster. What should I do ?
- If you don't care about ingressClass, or you have a lot of ingress objects without ingressClass configuration, you can run the ingress-controller with the flag `--watch-ingress-without-class=true`.

## What is the flag '--watch-ingress-without-class' ?
- Its a flag that is passed,as an argument, to the ingress-controller executable, in the pod spec. It looks like this ;
```
...
...
args:
  - /nginx-ingress-controller
  - --publish-service=$(POD_NAMESPACE)/ingress-nginx-dev-v1-test-controller
  - --election-id=ingress-controller-leader
  - --controller-class=k8s.io/ingress-nginx
  - --configmap=$(POD_NAMESPACE)/ingress-nginx-dev-v1-test-controller
  - --validating-webhook=:8443
  - --validating-webhook-certificate=/usr/local/certificates/cert
  - --validating-webhook-key=/usr/local/certificates/key
  - --watch-ingress-without-class=true
...
...
```

## I have more than one controller in my cluster and already use the annotation ?
No problem. This should still keep working, but we highly recommend you to test!

## I have more than one controller running in my cluster, and I want to use the new spec ?
In this scenario, you need to create multiple ingressClasses (see example one). But be aware that ingressClass works in a very specific way: you will need to change the .spec.controller value in your IngressClass and point the controller to the relevant ingressClass. Let's see some example, supposing that you have two Ingress Classes:

- Ingress-Nginx-IngressClass-1 with .spec.controller equals to "k8s.io/ingress-nginx1"
- Ingress-Nginx-IngressClass-2 with .spec.controller equals to "k8s.io/ingress-nginx2"
When deploying your ingress controllers, you will have to change the `--controller-class` field as follows:

Ingress-Nginx-Controller-nginx1 with `k8s.io/ingress-nginx1`
Ingress-Nginx-Controller-nginx2 with `k8s.io/ingress-nginx2`
Then, when you create an Ingress Object with IngressClassName = `ingress-nginx2`, it will look for controllers with `controller-class=k8s.io/ingress-nginx2` and as `Ingress-Nginx-Controller-nginx2` is watching objects that points to `ingressClass="k8s.io/ingress-nginx2`, it will serve that object, while `Ingress-Nginx-Controller-nginx1` will ignore the ingress object.

Bear in mind that, if your `Ingress-Nginx-Controller-nginx2` is started with the flag `--watch-ingress-without-class=true`, then it will serve ;
- objects without ingress-class
- objects with the annotation configured in flag `--ingress-class` and same class value
- and also objects pointing to the ingressClass that have the same .spec.controller as configured in `--controller-class`


## I am seeing this error message in the logs of the Ingress-NGINX controller "ingress class annotation is not equal to the expected by Ingress Controller". Why ?
- It is highly likely that you will also see the name of the ingress resource in the same error message. This error messsage has been observed on use the deprecated annotation, to spec the ingressClass, in a ingress resource manifest. It is recommended to use the ingress.spec.ingressClassName field, of the ingress resource, to spec the name of the ingressClass of the ingress resource being configured.
