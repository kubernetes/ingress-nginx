module k8s.io/ingress-nginx

go 1.13

require (
	github.com/armon/go-proxyproto v0.0.0-20200108142055-f0b8253b1507
	github.com/eapache/channels v1.1.0
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/json-iterator/go v1.1.9
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/go-ps v1.0.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moul/pb v0.0.0-20180404114147-54bdd96e6a52
	github.com/ncabatoff/process-exporter v0.6.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/opencontainers/runc v1.0.0-rc9
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.4.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/tallclair/mdtoc v0.0.0-20190627191617-4dc3d6f90813
	github.com/zakjan/cert-chain-resolver v0.0.0-20180703112424-6076e1ded272
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	google.golang.org/grpc v1.23.1
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/pool.v3 v3.1.1

	k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/apiserver v0.17.2
	k8s.io/cli-runtime v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/code-generator v0.17.2
	k8s.io/component-base v0.17.2
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.17.2
	moul.io/http2curl v1.0.0 // indirect
	pault.ag/go/sniff v0.0.0-20200207005214-cf7e4d167732
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	k8s.io/api => k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.2
	k8s.io/apiserver => k8s.io/apiserver v0.17.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.2
	k8s.io/client-go => k8s.io/client-go v0.17.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.2
	k8s.io/code-generator => k8s.io/code-generator v0.17.2
	k8s.io/component-base => k8s.io/component-base v0.17.2
	k8s.io/cri-api => k8s.io/cri-api v0.17.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.2
	k8s.io/kubectl => k8s.io/kubectl v0.17.2
	k8s.io/kubelet => k8s.io/kubelet v0.17.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.2
	k8s.io/metrics => k8s.io/metrics v0.17.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.2

)
