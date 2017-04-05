
## Backend-Service BalancingMode Updater
**For non-GKE Users**

Earlier versions of the GLBC created GCP BackendService resources with no balancing mode specified. By default the API used CPU UTILIZATION. The "internal load balancer" feature provided by GCP requires backend services to have the balancing mode RATE. In order to have a K8s cluster with an internal load balancer and ingress resources, you'll need to perform some manual steps.

#### Why
There are two GCP requirements that complicate changing the backend service balancing mode.
1. An instance can only belong to one loadbalancer instance group (a group that has at least one backend service pointing to it).
1. An load balancer instance group can only have one balancing mode for all the backend services pointing to it.

#### Complicating factors
1. You cannot atomically update a set of backend services to a new backend mode.
1. The default backend service in the `kube-system` namespace exists, so you'll have at least two backend services.

#### Your Options
- (UNTESTED) If you only have one service being referenced by ingresses AND that service is the default backend as specified in the Ingress spec. (resulting in one used backend service and one unused backend service)
   1. Go to the GCP Console
   1. Delete the kube-system's default backend service
   1. Change the balancing mode of the used backend service.   

 The GLBC should recreate the default backend service at its resync interval.


- Re-create all ingress resources. The GLBC will use RATE mode when it's not blocked by backend services with UTILIZATION mode.
  - Must be running GLBC version >0.9.1
  - Must delete all ingress resources before re-creating


- Run this updater tool

#### How the updater works
1. Create temporary instance groups `k8s-ig-migrate` in each zone where a `k8s-ig-{cluster_id}` exists.
1. Update all backend-services to point to both original and temporary instance groups (mode of the new backend doesn't matter)
1. Slowly migrate instances from original to temporary groups.
1. Update all backend-services to remove pointers to original instance groups.
1. Update all backend-services to point to original groups (with new balancing mode!)
1. Slowly migrate instances from temporary to original groups.
1. Update all backend-services to remove pointers to temporary instance groups.
1. Delete temporary instance groups

#### Required Testing
- [ ] Up time is not effected when switching instance groups
- [ ] An active GLBC does not negatively interfere with this updater

#### TODO
- [ ] Use GCE CloudProvider package in order to utilize the `waitForOp` functionality in order to remove some sleeps.
