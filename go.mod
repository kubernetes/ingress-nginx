module k8s.io/ingress-nginx

go 1.12

require (
	cloud.google.com/go v0.38.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000 // indirect
	github.com/armon/go-proxyproto v0.0.0-20190211145416-68259f75880e
	github.com/eapache/channels v1.1.0
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/swag v0.19.0 // indirect
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
	github.com/opencontainers/runc v0.1.1
	github.com/parnurzeal/gorequest v0.2.15
	github.com/paultag/sniff v0.0.0-20170624152000-87325c3dddf4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190115171406-56726106282f
	github.com/prometheus/common v0.2.0
	github.com/prometheus/procfs v0.0.0-20190328153300-af7bedc223fb // indirect
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
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/apiserver v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/klog v0.4.0
	k8s.io/kubernetes v1.15.1
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a // indirect
	sigs.k8s.io/controller-runtime v0.1.10
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1
	k8s.io/api => k8s.io/api v0.0.0-20190718183219-b59d8169aab5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085550-3a2f62f126c9
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190711103026-7bf792636534
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190718184206-a1aa83af71a7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085659-429467d76d0e
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190718183610-8e956561bbf5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090041-1a7e1f6630cd
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090010-a60497bb9ffa
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085131-4cd66be69262
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090114-816aa063c73d
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085316-c835efc41000
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085943-52018c8ce3c1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085811-cc0b23ba60a9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085909-5dfb14b3a101
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085837-98477dc0c87c
	k8s.io/kubernetes => k8s.io/kubernetes v1.15.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090159-a9e4f3cb5bf3
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085627-5b02f62e9559
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085357-8191e314a1f7
)
