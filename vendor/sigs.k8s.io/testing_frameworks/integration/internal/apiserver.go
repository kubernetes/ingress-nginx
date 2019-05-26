package internal

var APIServerDefaultArgs = []string{
	"--etcd-servers={{ if .EtcdURL }}{{ .EtcdURL.String }}{{ end }}",
	"--cert-dir={{ .CertDir }}",
	"--insecure-port={{ if .URL }}{{ .URL.Port }}{{ end }}",
	"--insecure-bind-address={{ if .URL }}{{ .URL.Hostname }}{{ end }}",
	"--secure-port={{ if .SecurePort }}{{ .SecurePort }}{{ end }}",
}

func DoAPIServerArgDefaulting(args []string) []string {
	if len(args) != 0 {
		return args
	}

	return APIServerDefaultArgs
}
