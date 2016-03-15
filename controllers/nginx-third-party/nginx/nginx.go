package nginx

// NGINXController Updates NGINX configuration, starts and reloads NGINX
type NGINXController struct {
	resolver       string
	nginxConfdPath string
	nginxCertsPath string
	local          bool
}

// IngressNGINXConfig describes an NGINX configuration
type IngressNGINXConfig struct {
	Upstreams []Upstream
	Servers   []Server
}

// Upstream describes an NGINX upstream
type Upstream struct {
	Name     string
	Backends []UpstreamServer
}

type UpstreamByNameServers []Upstream

func (c UpstreamByNameServers) Len() int      { return len(c) }
func (c UpstreamByNameServers) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c UpstreamByNameServers) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

// UpstreamServer describes a server in an NGINX upstream
type UpstreamServer struct {
	Address string
	Port    string
}

type UpstreamServerByAddrPort []UpstreamServer

func (c UpstreamServerByAddrPort) Len() int      { return len(c) }
func (c UpstreamServerByAddrPort) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c UpstreamServerByAddrPort) Less(i, j int) bool {
	iName := c[i].Address
	jName := c[j].Address
	if iName != jName {
		return iName < jName
	}

	iU := c[i].Port
	jU := c[j].Port
	return iU < jU
}

// Server describes an NGINX server
type Server struct {
	Name              string
	Locations         []Location
	SSL               bool
	SSLCertificate    string
	SSLCertificateKey string
}

type ServerByNamePort []Server

func (c ServerByNamePort) Len() int      { return len(c) }
func (c ServerByNamePort) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c ServerByNamePort) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

// Location describes an NGINX location
type Location struct {
	Path     string
	Upstream Upstream
}

type locByPathUpstream []Location

func (c locByPathUpstream) Len() int      { return len(c) }
func (c locByPathUpstream) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c locByPathUpstream) Less(i, j int) bool {
	return c[i].Path < c[j].Path
}

// NewUpstreamWithDefaultServer creates an upstream with the default server.
// proxy_pass to an upstream with the default server returns 502.
// We use it for services that have no endpoints
func NewUpstreamWithDefaultServer(name string) Upstream {
	return Upstream{
		Name:     name,
		Backends: []UpstreamServer{UpstreamServer{Address: "127.0.0.1", Port: "8181"}},
	}
}
