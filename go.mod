module k8s.io/ingress-nginx

go 1.12

require (
	cloud.google.com/go v0.37.2 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	github.com/Azure/go-autorest v11.7.1+incompatible // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/Sirupsen/logrus v1.4.0 // indirect
	github.com/armon/go-proxyproto v0.0.0-20190211145416-68259f75880e
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-units v0.3.3 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/eapache/channels v1.1.0
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190410145444-c548f45dcf1d // indirect
	github.com/emicklei/go-restful v2.9.4+incompatible // indirect
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa // indirect
	github.com/go-openapi/jsonpointer v0.19.0 // indirect
	github.com/go-openapi/jsonreference v0.19.0 // indirect
	github.com/go-openapi/spec v0.19.0 // indirect
	github.com/go-openapi/swag v0.19.0 // indirect
	github.com/gofortune/gofortune v0.0.1-snapshot // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/gophercloud/gophercloud v0.0.0-20190410012400-2c55d17f707c // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/json-iterator/go v1.1.6
	github.com/kisielk/errcheck v1.2.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/ncabatoff/go-seq v0.0.0-20180805175032-b08ef85ed833 // indirect
	github.com/ncabatoff/process-exporter v0.0.0-20180915144445-bdf24ef23850
	github.com/ncabatoff/procfs v0.0.0-20180903163354-e1a38cb53622 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/runc v0.1.1
	github.com/parnurzeal/gorequest v0.2.15
	github.com/paultag/sniff v0.0.0-20170624152000-87325c3dddf4
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.2.0
	github.com/prometheus/procfs v0.0.0-20190328153300-af7bedc223fb // indirect
	github.com/sirupsen/logrus v1.4.1 // indirect
	github.com/smartystreets/goconvey v0.0.0-20190330032615-68dc04aab96a // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/tv42/httpunix v0.0.0-20150427012821-b75d8614f926
	github.com/zakjan/cert-chain-resolver v0.0.0-20180703112424-6076e1ded272
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	golang.org/x/net v0.0.0-20190509222800-a4d6f7feada5 // indirect
	golang.org/x/sys v0.0.0-20190509141414-a5b02f93d862 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20190513172155-8696927a2839 // indirect
	google.golang.org/grpc v1.19.1
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/pool.v3 v3.1.1

	k8s.io/api v0.0.0-20190512063542-eae0ddcf85ba
	k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery v0.0.0-20190511063452-5b67e417bf61
	k8s.io/apiserver v0.0.0-20190313205120-8b27c41bdbb1
	k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/cloud-provider v0.0.0-20190323031113-9c9d72d1bf90 // indirect
	k8s.io/code-generator v0.0.0-00010101000000-000000000000
	k8s.io/component-base v0.0.0-20190313120452-4727f38490bc
	k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a // indirect
	k8s.io/klog v0.3.0
	k8s.io/kube-openapi v0.0.0-20190510232812-a01b7d5d6c22 // indirect
	k8s.io/kubernetes v1.14.1
	k8s.io/utils v0.0.0-20190308190857-21c4ce38f2a7 // indirect
	sigs.k8s.io/kind v0.0.0-20190510193412-0a6ceaec7a64 // indirect
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190511023357-639c964206c2
)
