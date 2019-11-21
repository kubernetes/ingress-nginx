module k8s.io/ingress-nginx

go 1.12

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/armon/go-proxyproto v0.0.0-20190211145416-68259f75880e
	github.com/eapache/channels v1.1.0
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.7
	github.com/json-iterator/go v1.1.7
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/moul/pb v0.0.0-20180404114147-54bdd96e6a52
	github.com/ncabatoff/go-seq v0.0.0-20180805175032-b08ef85ed833 // indirect
	github.com/ncabatoff/process-exporter v0.0.0-20180915144445-bdf24ef23850
	github.com/ncabatoff/procfs v0.0.0-20180903163354-e1a38cb53622 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/opencontainers/runc v1.0.0-rc2.0.20190611121236-6cc515888830
	github.com/parnurzeal/gorequest v0.2.15
	github.com/paultag/sniff v0.0.0-20170624152000-87325c3dddf4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190115171406-56726106282f
	github.com/prometheus/common v0.2.0
	github.com/prometheus/procfs v0.0.0-20190328153300-af7bedc223fb // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/tallclair/mdtoc v0.0.0-20190627191617-4dc3d6f90813
	github.com/zakjan/cert-chain-resolver v0.0.0-20180703112424-6076e1ded272
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc
	google.golang.org/grpc v1.23.0
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/pool.v3 v3.1.1

	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/apiserver v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/code-generator v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.16.0
	sigs.k8s.io/controller-runtime v0.1.10
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1
	k8s.io/api => k8s.io/api v0.0.0-20190919035539-41700d9d0c5b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918080820-40952ff8d5b6
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190917163033-a891081239f5
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918040322-b11291ff0a50
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190916161055-1f2b8882058b
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190916161804-3af65fe0d627
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190913091112-9859410eb5f6
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912042602-ebc0eb3a5c23
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918040032-61bc4cc48c91
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828121515-24ae4d4e8b03
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190913091657-9745ba0e69cf
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190917160357-d172b2afe66f
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190831080900-37642bccd2bd
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190831080623-67697732d2b9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190831080806-2937954be24b
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190918164019-21692a0861df
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190913090242-4a988d086279
	k8s.io/kubernetes => k8s.io/kubernetes v1.16.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190917162005-1c48f4e41cb3
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190913085057-ae59e7bc40d2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190913083406-d0cd75593f8b
)
