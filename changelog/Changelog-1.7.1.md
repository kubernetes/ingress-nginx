# Changelog

### 1.7.1
Images:

 * registry.k8s.io/ingress-nginx/controller:v1.7.1@sha256:7244b95ea47bddcb8267c1e625fb163fc183ef55448855e3ac52a7b260a60407
 * registry.k8s.io/ingress-nginx/controller-chroot:v1.7.1@sha256:e35d5ab487861b9d419c570e3530589229224a0762c7b4d2e2222434abb8d988
 
### All Changes:

* Update TAG - 1.7.1 (#9922)
* Update dependabot to watch docker images (#9600)
* [helm] Support custom port configuration for internal service (#9846)
* Add support for --container flag (#9703)
* Fix typo in OpenTelemetry (#9903)
* ensure make lua-test runs locally (#9902)
* update k8s.io dependecies to v0.26.4 (#9893)
* Adding resource type to default HPA configuration to resolve issues with Terraform helm chart usage (#9803)
* I have not been able to fulfill my maintainer responsibilities for a while already, making it official now. (#9883)
* Update k8s versions (#9879)
* README: Update `external-dns` link. (#9866)
* Fastcgi configmap should be on the same namespace of ingress (#9863)
* Deprecate and remove influxdb feature (#9861)
* Remove deprecated annotation secure-upstream (#9862)
* Exclude socket metrics (#9770)
* Chart: Improve `README.md`. (#9831)
* update all container tags with date and sha, upgrade all containers (#9834)
* updated NGINX_BASE image in project (#9829)
* ISO 8601 date format (#9682)
* Values: Fix indention of commented values. (#9812)
* The Ingress-Nginx project recently released version 1.7.0 of the controller, but the deployment documentation still referenced version 1.6.4. This commit updates the documentation to reference the latest version, ensuring that users have access to the most up-to-date information. Fixes#9787 (#9788)

### Dependencies updates: 
* Bump github.com/opencontainers/runc from 1.1.6 to 1.1.7 (#9912)
* Bump github.com/prometheus/client_golang from 1.14.0 to 1.15.0 (#9868)
* Bump aquasecurity/trivy-action from 0.9.2 to 0.10.0 (#9888)
* Bump github.com/opencontainers/runc from 1.1.5 to 1.1.6 (#9867)
* Bump actions/checkout from 3.5.0 to 3.5.2 (#9870)
* Bump golang.org/x/crypto from 0.7.0 to 0.8.0 (#9838)
* Bump github.com/spf13/cobra from 1.6.1 to 1.7.0 (#9839)
* Bump actions/add-to-project from 0.4.1 to 0.5.0 (#9840)
* Bump actions/checkout from 3.4.0 to 3.5.0 (#9798)
* Bump ossf/scorecard-action from 2.1.2 to 2.1.3 (#9823)
* Bump github.com/opencontainers/runc from 1.1.4 to 1.1.5 (#9806)
* Bump actions/stale from 7.0.0 to 8.0.0 (#9799)
* Bump rajatjindal/krew-release-bot from 0.0.43 to 0.0.46 (#9797)
* Bump actions/setup-go from 3.5.0 to 4.0.0 (#9796)
* Bump github.com/imdario/mergo from 0.3.13 to 0.3.15 (#9795)
* Bump google.golang.org/grpc from 1.53.0 to 1.54.0 (#9794)
* Bump sigs.k8s.io/controller-runtime from 0.14.5 to 0.14.6 (#9822)
 
**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-controller-v1.7.0...controller-controller-v1.7.1
