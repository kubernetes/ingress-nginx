
## (ALPHA) Backend-Service BalancingMode Updater
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

#### How to run
```shell
go run main.go {project-id} {region} {target-balance-mode}

#Examples
# for upgrading
go run main.go my-project us-central1 RATE

# for reversing
go run main.go my-project us-central1 UTILIZATION
```

**Example Run**
```shell
➜  go run mode-updater.go nicksardo-project us-central1 RATE    

Backend-Service BalancingMode Updater 0.1
Backend Services:
 -  k8s-be-31165--c4424dd5f02d3cad
 -  k8s-be-31696--c4424dd5f02d3cad
Instance Groups:
 - k8s-ig--c4424dd5f02d3cad (us-central1-a)

Step 1: Creating temporary instance groups in relevant zones
 - k8s-ig--migrate (us-central1-a)

Step 2: Update backend services to point to original and temporary instance groups
 - k8s-be-31165--c4424dd5f02d3cad
 - k8s-be-31696--c4424dd5f02d3cad

Step 3: Migrate instances to temporary group
 - kubernetes-minion-group-f060 (us-central1-a): removed from k8s-ig--c4424dd5f02d3cad, added to k8s-ig--migrate
 - kubernetes-minion-group-pnbl (us-central1-a): removed from k8s-ig--c4424dd5f02d3cad, added to k8s-ig--migrate
 - kubernetes-minion-group-t6dl (us-central1-a): removed from k8s-ig--c4424dd5f02d3cad, added to k8s-ig--migrate

Step 4: Update backend services to point only to temporary instance groups
 - k8s-be-31165--c4424dd5f02d3cad
 - k8s-be-31696--c4424dd5f02d3cad

Step 5: Update backend services to point to both temporary and original (with new balancing mode) instance groups
 - k8s-be-31165--c4424dd5f02d3cad
 - k8s-be-31696--c4424dd5f02d3cad

Step 6: Migrate instances back to original groups
 - kubernetes-minion-group-f060 (us-central1-a): removed from k8s-ig--migrate, added to k8s-ig--c4424dd5f02d3cad
 - kubernetes-minion-group-pnbl (us-central1-a): removed from k8s-ig--migrate, added to k8s-ig--c4424dd5f02d3cad
 - kubernetes-minion-group-t6dl (us-central1-a): removed from k8s-ig--migrate, added to k8s-ig--c4424dd5f02d3cad

Step 7: Update backend services to point only to original instance groups
 - k8s-be-31165--c4424dd5f02d3cad
 - k8s-be-31696--c4424dd5f02d3cad

Step 8: Delete temporary instance groups
 - k8s-ig--migrate (us-central1-a)
```

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
- [x] If only one backend-service exists, just update it in place.
- [x] If all backend-services are already the target balancing mode, early return.
- [ ] Wait for op completion instead of sleeping
- [ ] Adjust warning

#### Warning
This tool hasn't been fully tested. Use at your own risk.
