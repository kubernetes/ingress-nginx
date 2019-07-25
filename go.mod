module k8s.io/ingress-nginx

go 1.12

require (
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000 // indirect
	github.com/armon/go-proxyproto v0.0.0-20190211145416-68259f75880e
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/eapache/channels v1.1.0
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.7
	github.com/json-iterator/go v1.1.6
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/moul/pb v0.0.0-20180404114147-54bdd96e6a52
	github.com/ncabatoff/go-seq v0.0.0-20180805175032-b08ef85ed833 // indirect
	github.com/ncabatoff/process-exporter v0.0.0-20180915144445-bdf24ef23850
	github.com/ncabatoff/procfs v0.0.0-20180903163354-e1a38cb53622 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/runc v0.1.1
	github.com/parnurzeal/gorequest v0.2.15
	github.com/paultag/sniff v0.0.0-20170624152000-87325c3dddf4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190115171406-56726106282f
	github.com/prometheus/common v0.2.0
	github.com/prometheus/procfs v0.0.0-20190328153300-af7bedc223fb // indirect
	github.com/smartystreets/goconvey v0.0.0-20190710185942-9d28bd7c0945 // indirect
	github.com/spf13/cobra v0.0.4
	github.com/spf13/pflag v1.0.3
	github.com/tallclair/mdtoc v0.0.0-20190627191617-4dc3d6f90813
	github.com/tv42/httpunix v0.0.0-20150427012821-b75d8614f926
	github.com/zakjan/cert-chain-resolver v0.0.0-20180703112424-6076e1ded272
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980
	google.golang.org/grpc v1.19.1
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/pool.v3 v3.1.1

	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0-20190626210203-fdc73e13f9a6
	k8s.io/apimachinery v0.0.0
	k8s.io/apiserver v0.0.0-20190625052034-8c075cba2f8c
	k8s.io/cli-runtime v0.0.0-20190711111425-61e036b70227
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cloud-provider v0.0.0-20190711113108-0d51509e0ef5 // indirect
	k8s.io/code-generator v0.0.0-20190620073620-d55040311883
	k8s.io/component-base v0.0.0-20190626045757-ca439aa083f5
	k8s.io/klog v0.3.3
	k8s.io/kubernetes v0.0.0
	sigs.k8s.io/controller-runtime v0.1.10
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1

	k8s.io/api => k8s.io/api v0.0.0-20190703205437-39734b2a72fe
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190703205208-4cfb76a8bf76
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.3
)
