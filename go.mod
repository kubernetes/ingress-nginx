module k8s.io/ingress-nginx

go 1.14

require (
	github.com/armon/go-proxyproto v0.0.0-20200108142055-f0b8253b1507
	github.com/eapache/channels v1.1.0
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa // indirect
	github.com/gavv/httpexpect/v2 v2.1.0
	github.com/imdario/mergo v0.3.7
	github.com/json-iterator/go v1.1.10
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/go-ps v1.0.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.3.2
	github.com/moul/pb v0.0.0-20180404114147-54bdd96e6a52
	github.com/ncabatoff/process-exporter v0.7.1
	github.com/onsi/ginkgo v1.14.0
	github.com/opencontainers/runc v1.0.0-rc10
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.10.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tallclair/mdtoc v1.0.0
	github.com/zakjan/cert-chain-resolver v0.0.0-20200409100953-fa92b0b5236f
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7
	google.golang.org/grpc v1.26.0
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/pool.v3 v3.1.1
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/apiserver v0.18.5
	k8s.io/cli-runtime v0.18.5
	k8s.io/client-go v0.18.5
	k8s.io/code-generator v0.18.5
	k8s.io/component-base v0.18.5
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.18.5
	pault.ag/go/sniff v0.0.0-20200207005214-cf7e4d167732
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	k8s.io/api => k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.5
	k8s.io/apiserver => k8s.io/apiserver v0.18.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.5
	k8s.io/client-go => k8s.io/client-go v0.18.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.5
	k8s.io/code-generator => k8s.io/code-generator v0.18.5
	k8s.io/component-base => k8s.io/component-base v0.18.5
	k8s.io/cri-api => k8s.io/cri-api v0.18.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.5
	k8s.io/kubectl => k8s.io/kubectl v0.18.5
	k8s.io/kubelet => k8s.io/kubelet v0.18.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.5
	k8s.io/metrics => k8s.io/metrics v0.18.5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.5
)
