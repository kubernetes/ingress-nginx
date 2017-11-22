package file

const (
	// AuthDirectory default directory used to store files
	// to authenticate request
	AuthDirectory = "/etc/ingress-controller/auth"

	// DefaultSSLDirectory defines the location where the SSL certificates will be generated
	// This directory contains all the SSL certificates that are specified in Ingress rules.
	// The name of each file is <namespace>-<secret name>.pem. The content is the concatenated
	// certificate and key.
	DefaultSSLDirectory = "/ingress-controller/ssl"
)

var (
	directories = []string{
		"/etc/nginx/template",
		"/run",
		DefaultSSLDirectory,
		AuthDirectory,
	}

	files = []string{
		"/run/nginx.pid",
	}
)
