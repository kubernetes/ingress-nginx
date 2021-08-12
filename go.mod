module k8s.io/ingress-nginx

go 1.16

require (
	github.com/armon/go-proxyproto v0.0.0-20210323213023-7e956b284f0a
	github.com/eapache/channels v1.1.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gavv/httpexpect/v2 v2.3.1
	github.com/imdario/mergo v0.3.12
	github.com/json-iterator/go v1.1.11
	github.com/kylelemons/godebug v1.1.0
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mitchellh/go-ps v1.0.0
	github.com/mitchellh/hashstructure v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/moul/pb v0.0.0-20180404114147-54bdd96e6a52
	github.com/ncabatoff/process-exporter v0.7.5
	github.com/onsi/ginkgo v1.16.4
	github.com/opencontainers/runc v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.30.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/zakjan/cert-chain-resolver v0.0.0-20210427055340-87e10242a981
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	google.golang.org/grpc v1.38.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/pool.v3 v3.1.1
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/apiserver v0.21.3
	k8s.io/cli-runtime v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/code-generator v0.21.3
	k8s.io/component-base v0.21.3
	k8s.io/klog/v2 v2.10.0
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176 // indirect
	pault.ag/go/sniff v0.0.0-20200207005214-cf7e4d167732
	sigs.k8s.io/controller-runtime v0.9.5
	sigs.k8s.io/mdtoc v1.0.1
)
