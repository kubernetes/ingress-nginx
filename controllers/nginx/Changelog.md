Changelog

Changelog

### 0.9-beta.11

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.11`

Fixes NGINX [CVE-2017-7529](http://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-7529)

*Changes:*

- [X] [#659](https://github.com/kubernetes/ingress/pull/659) [nginx] TCP configmap should allow listen proxy_protocol per service
- [X] [#730](https://github.com/kubernetes/ingress/pull/730) Add support for add_headers
- [X] [#808](https://github.com/kubernetes/ingress/pull/808) HTTP->HTTPS redirect does not work with use-proxy-protocol: "true"
- [X] [#921](https://github.com/kubernetes/ingress/pull/921) Make proxy-real-ip-cidr a comma separated list
- [X] [#930](https://github.com/kubernetes/ingress/pull/930) Add support for proxy protocol in TCP services
- [X] [#933](https://github.com/kubernetes/ingress/pull/933) Lint code
- [X] [#937](https://github.com/kubernetes/ingress/pull/937) Fix lint code errors
- [X] [#940](https://github.com/kubernetes/ingress/pull/940) Sets parameters for a shared memory zone of limit_conn_zone
- [X] [#949](https://github.com/kubernetes/ingress/pull/949) fix nginx version to 1.13.3 to fix integer overflow
- [X] [#956](https://github.com/kubernetes/ingress/pull/956) Simplify handling of ssl certificates
- [X] [#958](https://github.com/kubernetes/ingress/pull/958) Release ubuntu-slim:0.13
- [X] [#959](https://github.com/kubernetes/ingress/pull/959) Release nginx-slim 0.21
- [X] [#960](https://github.com/kubernetes/ingress/pull/960) Update nginx in ingress controller
- [X] [#964](https://github.com/kubernetes/ingress/pull/964) Support for proxy_headers_hash_bucket_size and proxy_headers_hash_max_size
- [X] [#966](https://github.com/kubernetes/ingress/pull/966) Fix error checking for pod name & NS
- [X] [#967](https://github.com/kubernetes/ingress/pull/967) Fix runningAddresses typo
- [X] [#968](https://github.com/kubernetes/ingress/pull/968) Fix missing hyphen in yaml for nginx RBAC example
- [X] [#973](https://github.com/kubernetes/ingress/pull/973) check number of servers in configuration comparator


### 0.9-beta.10

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.10`

Fix release 0.9-beta.9

### 0.9-beta.9

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.9`

*New Features:*

- Add support for arm and ppc64le


*Changes:*

- [X] [#548](https://github.com/kubernetes/ingress/pull/548) nginx: support multidomain certificates
- [X] [#620](https://github.com/kubernetes/ingress/pull/620) [nginx] Listening ports are not configurable, so ingress can't be run multiple times per node when using CNI
- [X] [#648](https://github.com/kubernetes/ingress/pull/648) publish-service argument isn't honored when ELB is internal only facing.
- [X] [#833](https://github.com/kubernetes/ingress/pull/833) WIP: Avoid reloads implementing Equals in structs
- [X] [#838](https://github.com/kubernetes/ingress/pull/838) Feature request: Add ingress annotation to enable upstream "keepalive" option
- [X] [#844](https://github.com/kubernetes/ingress/pull/844) ingress annotations affinity is not working
- [X] [#862](https://github.com/kubernetes/ingress/pull/862) Avoid reloads implementing Equaler interface
- [X] [#864](https://github.com/kubernetes/ingress/pull/864) Remove dead code
- [X] [#868](https://github.com/kubernetes/ingress/pull/868) Lint nginx code
- [X] [#871](https://github.com/kubernetes/ingress/pull/871) Add feature to allow sticky sessions per location
- [X] [#873](https://github.com/kubernetes/ingress/pull/873) Update README.md
- [X] [#876](https://github.com/kubernetes/ingress/pull/876) Add information about nginx controller flags
- [X] [#878](https://github.com/kubernetes/ingress/pull/878) Update go to 1.8.3
- [X] [#881](https://github.com/kubernetes/ingress/pull/881) Option to not remove loadBalancer status record?
- [X] [#882](https://github.com/kubernetes/ingress/pull/882) Add flag to skip the update of Ingress status on shutdown
- [X] [#885](https://github.com/kubernetes/ingress/pull/885) Don't use $proxy_protocol var which may be undefined.
- [X] [#886](https://github.com/kubernetes/ingress/pull/886) Add support for SubjectAltName in SSL certificates
- [X] [#888](https://github.com/kubernetes/ingress/pull/888) Update nginx-slim to 0.19
- [X] [#889](https://github.com/kubernetes/ingress/pull/889) Add PHOST to backend
- [X] [#890](https://github.com/kubernetes/ingress/pull/890) Improve variable configuration for source IP address
- [X] [#892](https://github.com/kubernetes/ingress/pull/892) Add upstream keepalive connections cache
- [X] [#897](https://github.com/kubernetes/ingress/pull/897) Update outdated ingress resource link
- [X] [#898](https://github.com/kubernetes/ingress/pull/898) add error check right when reload nginx fail
- [X] [#899](https://github.com/kubernetes/ingress/pull/899) Fix nginx error check
- [X] [#900](https://github.com/kubernetes/ingress/pull/900) After #862 changes in the configmap do not trigger a reload
- [X] [#901](https://github.com/kubernetes/ingress/pull/901) [doc] Update NGinX status port to 18080
- [X] [#902](https://github.com/kubernetes/ingress/pull/902) Always reload after a change in the configuration
- [X] [#904](https://github.com/kubernetes/ingress/pull/904) Fix nginx sticky sessions
- [X] [#906](https://github.com/kubernetes/ingress/pull/906) Fix race condition with closed channels
- [X] [#907](https://github.com/kubernetes/ingress/pull/907) nginx/proxy: allow specifying next upstream behaviour
- [X] [#910](https://github.com/kubernetes/ingress/pull/910) Feature request: use `X-Forwarded-Host` from the reverse proxy before
- [X] [#911](https://github.com/kubernetes/ingress/pull/911) Improve X-Forwarded-Host support
- [X] [#915](https://github.com/kubernetes/ingress/pull/915) Release nginx-slim 0.20
- [X] [#916](https://github.com/kubernetes/ingress/pull/916) Add arm and ppc64le support
- [X] [#919](https://github.com/kubernetes/ingress/pull/919) Apply the 'ssl-redirect' annotation per-location
- [X] [#922](https://github.com/kubernetes/ingress/pull/922) Add example of TLS termination using a classic ELB

### 0.9-beta.8

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.8`

*Changes:*

- [X] [#761](https://github.com/kubernetes/ingress/pull/761) NGINX TCP Ingresses do not bind on IPv6
- [X] [#850](https://github.com/kubernetes/ingress/pull/850) Fix IPv6 UDP stream section
- [X] [#851](https://github.com/kubernetes/ingress/pull/851) ensure private key and certificate match
- [X] [#852](https://github.com/kubernetes/ingress/pull/852) Don't expose certificate metrics for default server
- [X] [#846](https://github.com/kubernetes/ingress/pull/846) Match ServicePort to Endpoints by Name
- [X] [#854](https://github.com/kubernetes/ingress/pull/854) Document log-format-stream and log-format-upstream
- [X] [#847](https://github.com/kubernetes/ingress/pull/847) fix semicolon
- [X] [#848](https://github.com/kubernetes/ingress/pull/848) Add metric "ssl certificate expiration"
- [X] [#839](https://github.com/kubernetes/ingress/pull/839) "No endpoints" issue
- [X] [#845](https://github.com/kubernetes/ingress/pull/845) Fix no endpoints issue when named ports are used
- [X] [#822](https://github.com/kubernetes/ingress/pull/822) Release ubuntu-slim 0.11
- [X] [#824](https://github.com/kubernetes/ingress/pull/824) Update nginx-slim to 0.18
- [X] [#823](https://github.com/kubernetes/ingress/pull/823) Release nginx-slim 0.18
- [X] [#827](https://github.com/kubernetes/ingress/pull/827) Introduce working example of nginx controller with rbac
- [X] [#835](https://github.com/kubernetes/ingress/pull/835) Make log format json escaping configurable
- [X] [#843](https://github.com/kubernetes/ingress/pull/843) Avoid setting maximum number of open file descriptors lower than 1024
- [X] [#837](https://github.com/kubernetes/ingress/pull/837) Cleanup interface
- [X] [#836](https://github.com/kubernetes/ingress/pull/836) Make log format json escaping configurable
- [X] [#828](https://github.com/kubernetes/ingress/pull/828) Wrap IPv6 endpoints in []
- [X] [#821](https://github.com/kubernetes/ingress/pull/821) nginx-ingress: occasional 503 Service Temporarily Unavailable
- [X] [#829](https://github.com/kubernetes/ingress/pull/829) feat(template): wrap IPv6 addresses in []
- [X] [#786](https://github.com/kubernetes/ingress/pull/786) Update echoserver image version in examples
- [X] [#825](https://github.com/kubernetes/ingress/pull/825) Create or delete ingress based on class annotation
- [X] [#790](https://github.com/kubernetes/ingress/pull/790) #789 removing duplicate X-Real-IP header 
- [X] [#792](https://github.com/kubernetes/ingress/pull/792) Avoid checking if the controllers are synced
- [X] [#798](https://github.com/kubernetes/ingress/pull/798) nginx: RBAC for leader election
- [X] [#799](https://github.com/kubernetes/ingress/pull/799) could not build variables_hash
- [X] [#809](https://github.com/kubernetes/ingress/pull/809) Fix dynamic variable name
- [X] [#804](https://github.com/kubernetes/ingress/pull/804) Fix #798 - RBAC for leader election
- [X] [#806](https://github.com/kubernetes/ingress/pull/806) fix ingress rbac roles
- [X] [#811](https://github.com/kubernetes/ingress/pull/811) external auth - proxy_pass_request_body off + big bodies give 500/413
- [X] [#785](https://github.com/kubernetes/ingress/pull/785) Publish echoheader image
- [X] [#813](https://github.com/kubernetes/ingress/pull/813) Added client_max_body_size to authPath location
- [X] [#814](https://github.com/kubernetes/ingress/pull/814) rbac-nginx: resourceNames cannot filter create verb
- [X] [#774](https://github.com/kubernetes/ingress/pull/774) Add IPv6 support in TCP and UDP stream section
- [X] [#784](https://github.com/kubernetes/ingress/pull/784) Allow customization of variables hash tables
- [X] [#782](https://github.com/kubernetes/ingress/pull/782) Set "proxy_pass_header Server;"
- [X] [#783](https://github.com/kubernetes/ingress/pull/783) nginx/README.md: clarify app-root and fix example hyperlink
- [X] [#787](https://github.com/kubernetes/ingress/pull/787) Add setting to allow returning the Server header from the backend

### 0.9-beta.7

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.7`

*Changes:*

- [X] [#777](https://github.com/kubernetes/ingress/pull/777) Update sniff parser to fix index out of bound error 

### 0.9-beta.6

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.6`

*Changes:*

- [X] [#647](https://github.com/kubernetes/ingress/pull/647) ingress.class enhancement for debugging.
- [X] [#708](https://github.com/kubernetes/ingress/pull/708) ingress losing real source IP when tls enabled
- [X] [#760](https://github.com/kubernetes/ingress/pull/760) Change recorder event scheme
- [X] [#704](https://github.com/kubernetes/ingress/pull/704) fix nginx reload flags '-c'
- [X] [#757](https://github.com/kubernetes/ingress/pull/757) Replace use of endpoints as locks with configmap
- [X] [#752](https://github.com/kubernetes/ingress/pull/752) nginx ingress header config backwards
- [X] [#756](https://github.com/kubernetes/ingress/pull/756) Fix bad variable assignment in template nginx
- [X] [#729](https://github.com/kubernetes/ingress/pull/729) Release nginx-slim 0.17
- [X] [#755](https://github.com/kubernetes/ingress/pull/755) Fix server name hash maxSize default value
- [X] [#741](https://github.com/kubernetes/ingress/pull/741) Update golang dependencies
- [X] [#749](https://github.com/kubernetes/ingress/pull/749) Remove service annotation for namedPorts
- [X] [#740](https://github.com/kubernetes/ingress/pull/740) Refactoring whitelist source IP verification
- [X] [#734](https://github.com/kubernetes/ingress/pull/734) Specify nginx image arch
- [X] [#728](https://github.com/kubernetes/ingress/pull/728) Update nginx image
- [X] [#723](https://github.com/kubernetes/ingress/pull/723) update readme about vts metrics
- [X] [#726](https://github.com/kubernetes/ingress/pull/726) Release ubuntu-slim 0.10
- [X] [#727](https://github.com/kubernetes/ingress/pull/727) [nginx] whitelist-source-range doesn’t work on ssl port
- [X] [#709](https://github.com/kubernetes/ingress/pull/709) Add config for X-Forwarded-For trust
- [X] [#679](https://github.com/kubernetes/ingress/pull/679) add getenv
- [X] [#680](https://github.com/kubernetes/ingress/pull/680) nginx/pkg/config: delete unuseful variable
- [X] [#716](https://github.com/kubernetes/ingress/pull/716) Add secure-verify-ca-secret annotation
- [X] [#722](https://github.com/kubernetes/ingress/pull/722) Remove go-reap and use tini as process reaper
- [X] [#725](https://github.com/kubernetes/ingress/pull/725) Add keepalive_requests and client_body_buffer_size options
- [X] [#724](https://github.com/kubernetes/ingress/pull/724) change the directory of default-backend.yaml
- [X] [#656](https://github.com/kubernetes/ingress/pull/656) Nginx Ingress Controller - Specify load balancing method
- [X] [#717](https://github.com/kubernetes/ingress/pull/717) delete unuseful variable
- [X] [#712](https://github.com/kubernetes/ingress/pull/712) Set $proxy_upstream_name before location directive
- [X] [#715](https://github.com/kubernetes/ingress/pull/715) Corrected annotation ex `signin-url` to `auth-url`
- [X] [#718](https://github.com/kubernetes/ingress/pull/718) nodeController sync
- [X] [#694](https://github.com/kubernetes/ingress/pull/694) SSL-Passthrough broken in beta.5
- [X] [#678](https://github.com/kubernetes/ingress/pull/678) Convert CN SSL Certificate to lowercase before comparison
- [X] [#690](https://github.com/kubernetes/ingress/pull/690) Fix IP in logs for https traffic
- [X] [#673](https://github.com/kubernetes/ingress/pull/673) Override load balancer alg view config map
- [X] [#675](https://github.com/kubernetes/ingress/pull/675) Use proxy-protocol to pass through source IP to nginx
- [X] [#707](https://github.com/kubernetes/ingress/pull/707) use nginx vts module version 0.1.14
- [X] [#702](https://github.com/kubernetes/ingress/pull/702) Document passing of ssl_client_cert to backend
- [X] [#688](https://github.com/kubernetes/ingress/pull/688) Add example of UDP loadbalancing
- [X] [#696](https://github.com/kubernetes/ingress/pull/696) [nginx] pass non-SNI TLS hello to default backend, Fixes #693
- [X] [#685](https://github.com/kubernetes/ingress/pull/685) Fix error in generated nginx.conf for optional hsts-preload


### 0.9-beta.5

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.5`

*Changes:*

- [X] [#663](https://github.com/kubernetes/ingress/pull/663) Remove helper required in go < 1.8
- [X] [#662](https://github.com/kubernetes/ingress/pull/662) Add debug information about ingress class
- [X] [#661](https://github.com/kubernetes/ingress/pull/661) Avoid running nginx if the configuration file is empty  
- [X] [#660](https://github.com/kubernetes/ingress/pull/660) Rollback queue refactoring  
- [X] [#654](https://github.com/kubernetes/ingress/pull/654) Update go version to 1.8


### 0.9-beta.4

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.4`

*New Features:*

- Add support for services of type ExternalName


*Changes:*

- [X] [#635](https://github.com/kubernetes/ingress/pull/635) Allow configuration of features underscores_in_headers and ignore_invalid_headers
- [X] [#633](https://github.com/kubernetes/ingress/pull/633) Fix lint errors
- [X] [#630](https://github.com/kubernetes/ingress/pull/630) Add example of TCP loadbalancing
- [X] [#629](https://github.com/kubernetes/ingress/pull/629) Add support for services of type ExternalName
- [X] [#624](https://github.com/kubernetes/ingress/pull/624) Compute server_names_hash_bucket_size correctly
- [X] [#615](https://github.com/kubernetes/ingress/pull/615) Process exited cleanly before we hit wait4
- [X] [#614](https://github.com/kubernetes/ingress/pull/614) Refactor nginx ssl passthrough
- [X] [#613](https://github.com/kubernetes/ingress/pull/613) Status leader election must consired the ingress class
- [X] [#607](https://github.com/kubernetes/ingress/pull/607) Allow custom server_names_hash_max_size & server_names_hash_bucket_size
- [X] [#601](https://github.com/kubernetes/ingress/pull/601) add a judgment
- [X] [#601](https://github.com/kubernetes/ingress/pull/600) Replace custom child reap code with go-reap
- [X] [#597](https://github.com/kubernetes/ingress/pull/599) Add flag to force namespace isolation
- [X] [#595](https://github.com/kubernetes/ingress/pull/595) Remove Host header from auth_request proxy configuration
- [X] [#588](https://github.com/kubernetes/ingress/pull/588) Read resolv.conf file just once
- [X] [#586](https://github.com/kubernetes/ingress/pull/586) Updated instructions to create an ingress controller build
- [X] [#583](https://github.com/kubernetes/ingress/pull/583) fixed lua_package_path in nginx.tmpl 
- [X] [#580](https://github.com/kubernetes/ingress/pull/580) Updated faq for running multiple ingress controller
- [X] [#579](https://github.com/kubernetes/ingress/pull/579) Detect if the ingress controller is running with multiple replicas
- [X] [#578](https://github.com/kubernetes/ingress/pull/578) Set different listeners per protocol version
- [X] [#577](https://github.com/kubernetes/ingress/pull/577) Avoid zombie child processes
- [X] [#576](https://github.com/kubernetes/ingress/pull/576) Replace secret workqueue
- [X] [#568](https://github.com/kubernetes/ingress/pull/568) Revert merge annotations to the implicit root context 
- [X] [#563](https://github.com/kubernetes/ingress/pull/563) Add option to disable hsts preload
- [X] [#560](https://github.com/kubernetes/ingress/pull/560) Fix intermittent misconfiguration of backend.secure and SessionAffinity
- [X] [#556](https://github.com/kubernetes/ingress/pull/556) Update nginx version and remove dumb-init
- [X] [#551](https://github.com/kubernetes/ingress/pull/551) Build namespace and ingress class as label
- [X] [#546](https://github.com/kubernetes/ingress/pull/546) Fix a couple of 'does not contains' typos
- [X] [#542](https://github.com/kubernetes/ingress/pull/542) Fix lint errors
- [X] [#540](https://github.com/kubernetes/ingress/pull/540) Add Backends.SSLPassthrough attribute
- [X] [#539](https://github.com/kubernetes/ingress/pull/539) Migrate to client-go
- [X] [#536](https://github.com/kubernetes/ingress/pull/536) add unit test cases for core/pkg/ingress/controller/backend_ssl
- [X] [#535](https://github.com/kubernetes/ingress/pull/535) Add test for ingress status update
- [X] [#532](https://github.com/kubernetes/ingress/pull/532) Add setting to configure ecdh curve
- [X] [#531](https://github.com/kubernetes/ingress/pull/531) Fix link to examples
- [X] [#530](https://github.com/kubernetes/ingress/pull/530) Fix link to custom nginx configuration
- [X] [#528](https://github.com/kubernetes/ingress/pull/528) Add reference to apiserver-host flag
- [X] [#527](https://github.com/kubernetes/ingress/pull/527) Add annotations to location of default backend (root context)
- [X] [#525](https://github.com/kubernetes/ingress/pull/525) Avoid negative values configuring the max number of open files
- [X] [#523](https://github.com/kubernetes/ingress/pull/523) Fix a typo in an error message
- [X] [#521](https://github.com/kubernetes/ingress/pull/521) nginx-ingress-controller is built twice by docker-build target
- [X] [#517](https://github.com/kubernetes/ingress/pull/517) Use whitelist-source-range from configmap when no annotation on ingress
- [X] [#516](https://github.com/kubernetes/ingress/pull/516) Convert WorkerProcesses setting to string to allow the value auto
- [X] [#512](https://github.com/kubernetes/ingress/pull/512) Fix typos regarding the ssl-passthrough annotation documentation
- [X] [#505](https://github.com/kubernetes/ingress/pull/505) add unit test cases for core/pkg/ingress/controller/annotations
- [X] [#503](https://github.com/kubernetes/ingress/pull/503) Add example for nginx in aws
- [X] [#502](https://github.com/kubernetes/ingress/pull/502) Add information about SSL Passthrough annotation 
- [X] [#500](https://github.com/kubernetes/ingress/pull/500) Improve TLS secret configuration
- [X] [#498](https://github.com/kubernetes/ingress/pull/498) Proper enqueue a secret on the secret queue
- [X] [#493](https://github.com/kubernetes/ingress/pull/493) Update nginx and vts module
- [X] [#490](https://github.com/kubernetes/ingress/pull/490) Add unit test case for named_port
- [X] [#488](https://github.com/kubernetes/ingress/pull/488) Adds support for CORS on error responses and Authorization header
- [X] [#485](https://github.com/kubernetes/ingress/pull/485) Fix typo nginx configMap vts metrics customization
- [X] [#481](https://github.com/kubernetes/ingress/pull/481) Remove unnecessary quote in nginx log format
- [X] [#471](https://github.com/kubernetes/ingress/pull/471) prometheus scrape annotations
- [X] [#460](https://github.com/kubernetes/ingress/pull/460) add example of 'run multiple haproxy ingress controllers as a deployment' 
- [X] [#459](https://github.com/kubernetes/ingress/pull/459) Add information about SSL certificates in the default log level
- [X] [#456](https://github.com/kubernetes/ingress/pull/456) Avoid upstreams with multiple servers with the same port
- [X] [#454](https://github.com/kubernetes/ingress/pull/454) Pass request port to real server
- [X] [#450](https://github.com/kubernetes/ingress/pull/450) fix nginx-tcp-and-udp on same port
- [X] [#446](https://github.com/kubernetes/ingress/pull/446) remove configmap validations
- [X] [#445](https://github.com/kubernetes/ingress/pull/445) Remove snakeoil certificate generation
- [X] [#442](https://github.com/kubernetes/ingress/pull/442) Fix a few bugs in the nginx-ingress-controller Makefile
- [X] [#441](https://github.com/kubernetes/ingress/pull/441) skip validation when configmap is empty
- [X] [#439](https://github.com/kubernetes/ingress/pull/439) Avoid a nil-reference when the temporary file cannot be created
- [X] [#438](https://github.com/kubernetes/ingress/pull/438) Improve English in error messages
- [X] [#437](https://github.com/kubernetes/ingress/pull/437) Reference constant


### 0.9-beta.3

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.3`

*New Features:*

- Custom log formats using `log-format-upstream` directive in the configuration configmap.
- Force redirect to SSL using the annotation `ingress.kubernetes.io/force-ssl-redirect`
- Prometheus metric for VTS status module (transparent, just enable vts stats)
- Improved external authentication adding `ingress.kubernetes.io/auth-signin` annotation. Please check this [example](https://github.com/kubernetes/ingress/tree/master/examples/external-auth/nginx)


*Breaking changes:*

- `ssl-dh-param` configuration in configmap is now the name of a secret that contains the Diffie-Hellman key

*Changes:*

- [X] [#433](https://github.com/kubernetes/ingress/pull/433) close over the ingress variable or the last assignment will be used 
- [X] [#424](https://github.com/kubernetes/ingress/pull/424) Manually sync secrets from certificate authentication annotations 
- [X] [#423](https://github.com/kubernetes/ingress/pull/423) Scrap json metrics from nginx vts module when enabled 
- [X] [#418](https://github.com/kubernetes/ingress/pull/418) Only update Ingress status for the configured class     
- [X] [#415](https://github.com/kubernetes/ingress/pull/415) Improve external authentication docs 
- [X] [#410](https://github.com/kubernetes/ingress/pull/410) Add support for "signin url" 
- [X] [#409](https://github.com/kubernetes/ingress/pull/409) Allow custom http2 header sizes 
- [X] [#408](https://github.com/kubernetes/ingress/pull/408) Review docs 
- [X] [#406](https://github.com/kubernetes/ingress/pull/406) Add debug info and fix spelling 
- [X] [#402](https://github.com/kubernetes/ingress/pull/402) allow specifying custom dh param 
- [X] [#397](https://github.com/kubernetes/ingress/pull/397) Fix external auth
- [X] [#394](https://github.com/kubernetes/ingress/pull/394) Update README.md 
- [X] [#392](https://github.com/kubernetes/ingress/pull/392) Fix http2 header size
- [X] [#391](https://github.com/kubernetes/ingress/pull/391) remove tmp nginx-diff files 
- [X] [#390](https://github.com/kubernetes/ingress/pull/390) Fix RateLimit comment 
- [X] [#385](https://github.com/kubernetes/ingress/pull/385) add Copyright 
- [X] [#382](https://github.com/kubernetes/ingress/pull/382) Ingress Fake Certificate generation 
- [X] [#380](https://github.com/kubernetes/ingress/pull/380) Fix custom log format 
- [X] [#373](https://github.com/kubernetes/ingress/pull/373) Cleanup 
- [X] [#371](https://github.com/kubernetes/ingress/pull/371) add configuration to disable listening on ipv6 
- [X] [#370](https://github.com/kubernetes/ingress/pull/270) Add documentation for ingress.kubernetes.io/force-ssl-redirect 
- [X] [#369](https://github.com/kubernetes/ingress/pull/369) Minor text fix for "ApiServer" 
- [X] [#367](https://github.com/kubernetes/ingress/pull/367) BuildLogFormatUpstream was always using the default log-format
- [X] [#366](https://github.com/kubernetes/ingress/pull/366) add_judgment 
- [X] [#365](https://github.com/kubernetes/ingress/pull/365) add ForceSSLRedirect ingress annotation 
- [X] [#364](https://github.com/kubernetes/ingress/pull/364) Fix error caused by increasing proxy_buffer_size (#363) 
- [X] [#362](https://github.com/kubernetes/ingress/pull/362) Fix ingress class 
- [X] [#360](https://github.com/kubernetes/ingress/pull/360) add example of 'run multiple nginx ingress controllers as a deployment' 
- [X] [#358](https://github.com/kubernetes/ingress/pull/358) Checks if the TLS secret contains a valid keypair structure 
- [X] [#356](https://github.com/kubernetes/ingress/pull/356) Disable listen only on ipv6 and fix proxy_protocol 
- [X] [#354](https://github.com/kubernetes/ingress/pull/354) add judgment 
- [X] [#352](https://github.com/kubernetes/ingress/pull/352) Add ability to customize upstream and stream log format 
- [X] [#351](https://github.com/kubernetes/ingress/pull/351) Enable custom election id for status sync. 
- [X] [#347](https://github.com/kubernetes/ingress/pull/347) Fix client source IP address 
- [X] [#345](https://github.com/kubernetes/ingress/pull/345) Fix lint error
- [X] [#344](https://github.com/kubernetes/ingress/pull/344) Refactoring of TCP and UDP services 
- [X] [#343](https://github.com/kubernetes/ingress/pull/343) Fix node lister when --watch-namespace is used 
- [X] [#341](https://github.com/kubernetes/ingress/pull/341) Do not run coverage check in the default target. 
- [X] [#340](https://github.com/kubernetes/ingress/pull/340) Add support for specify proxy cookie path/domain 
- [X] [#337](https://github.com/kubernetes/ingress/pull/337) Fix for formatting error introduced in #304 
- [X] [#335](https://github.com/kubernetes/ingress/pull/335) Fix for vet complaints: 
- [X] [#332](https://github.com/kubernetes/ingress/pull/332) Add annotation to customize nginx configuration 
- [X] [#331](https://github.com/kubernetes/ingress/pull/331) Correct spelling mistake 
- [X] [#328](https://github.com/kubernetes/ingress/pull/328) fix misspell "affinity" in main.go 
- [X] [#326](https://github.com/kubernetes/ingress/pull/326) add nginx daemonset example 
- [X] [#311](https://github.com/kubernetes/ingress/pull/311) Sort stream service ports to avoid extra reloads 
- [X] [#307](https://github.com/kubernetes/ingress/pull/307) Add docs for body-size annotation
- [X] [#306](https://github.com/kubernetes/ingress/pull/306) modify nginx readme 
- [X] [#304](https://github.com/kubernetes/ingress/pull/304) change 'buildSSPassthrouthUpstreams' to 'buildSSLPassthroughUpstreams' 


### 0.9-beta.2

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.2`

*New Features:*

- New configuration flag `proxy-set-headers` to allow set custom headers before send traffic to backends. [Example here](https://github.com/kubernetes/ingress/tree/master/examples/customization/custom-headers/nginx)
- Disable directive access_log globally using `disable-access-log: "true"` in the configuration ConfigMap.
- Sticky session per Ingress rule using the annotation `ingress.kubernetes.io/affinity`. [Example here](https://github.com/kubernetes/ingress/tree/master/examples/affinity/cookie/nginx)

*Changes:*

- [X] [#300](https://github.com/kubernetes/ingress/pull/300) Change nginx variable to use in filter of access_log
- [X] [#296](https://github.com/kubernetes/ingress/pull/296) Fix rewrite regex to match the start of the URL and not a substring
- [X] [#293](https://github.com/kubernetes/ingress/pull/293) Update makefile gcloud docker command
- [X] [#290](https://github.com/kubernetes/ingress/pull/290) Update nginx version in ingress controller to 1.11.10
- [X] [#286](https://github.com/kubernetes/ingress/pull/286) Add logs to help debugging and simplify default upstream configuration
- [X] [#285](https://github.com/kubernetes/ingress/pull/285) Added a Node StoreLister type
- [X] [#281](https://github.com/kubernetes/ingress/pull/281) Add chmod up directory tree for world read/execute on directories
- [X] [#279](https://github.com/kubernetes/ingress/pull/279) fix wrong link in the file of examples/README.md
- [X] [#275](https://github.com/kubernetes/ingress/pull/275) Pass headers to custom error backend
- [X] [#272](https://github.com/kubernetes/ingress/pull/272) Fix error getting class information from Ingress annotations
- [X] [#268](https://github.com/kubernetes/ingress/pull/268) minor: Fix typo in nginx README
- [X] [#265](https://github.com/kubernetes/ingress/pull/265) Fix rewrite annotation parser
- [X] [#262](https://github.com/kubernetes/ingress/pull/262) Add nginx README and configuration docs back
- [X] [#261](https://github.com/kubernetes/ingress/pull/261) types.go: fix typo in godoc
- [X] [#258](https://github.com/kubernetes/ingress/pull/258) Nginx sticky annotations
- [X] [#255](https://github.com/kubernetes/ingress/pull/255) Adds support for disabling access_log globally
- [X] [#247](https://github.com/kubernetes/ingress/pull/247) Fix wrong URL in nginx ingress configuration
- [X] [#246](https://github.com/kubernetes/ingress/pull/246) Add support for custom proxy headers using a ConfigMap
- [X] [#244](https://github.com/kubernetes/ingress/pull/244) Add information about cors annotation
- [X] [#241](https://github.com/kubernetes/ingress/pull/241) correct a spell mistake
- [X] [#232](https://github.com/kubernetes/ingress/pull/232) Change searchs with searches
- [X] [#231](https://github.com/kubernetes/ingress/pull/231) Add information about proxy_protocol in port 442
- [X] [#228](https://github.com/kubernetes/ingress/pull/228) Fix worker check issue
- [X] [#227](https://github.com/kubernetes/ingress/pull/227) proxy_protocol on ssl_passthrough listener
- [X] [#223](https://github.com/kubernetes/ingress/pull/223) Fix panic if a tempfile cannot be created
- [X] [#220](https://github.com/kubernetes/ingress/pull/220) Fixes for minikube usage instructions.
- [X] [#219](https://github.com/kubernetes/ingress/pull/219) Fix typo, add a couple of links. 
- [X] [#218](https://github.com/kubernetes/ingress/pull/218) Improve links from CONTRIBUTING.
- [X] [#217](https://github.com/kubernetes/ingress/pull/217) Fix an e2e link. 
- [X] [#212](https://github.com/kubernetes/ingress/pull/212) Simplify code to obtain TCP or UDP services
- [X] [#208](https://github.com/kubernetes/ingress/pull/208) Fix nil HTTP field
- [X] [#198](https://github.com/kubernetes/ingress/pull/198) Add an example for static-ip and deployment


### 0.9-beta.1

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.1`

*New Features:*

- SSL Passthrough
- New Flag `--publish-service` that set the Service fronting the ingress controllers
- Ingress status shows the correct IP/hostname address without duplicates
- Custom body sizes per Ingress
- Prometheus metrics


*Breaking changes:*

- Flag `--nginx-configmap` was replaced with `--configmap`
- Configmap field `body-size` was replaced with `proxy-body-size`

*Changes:*

- [X] [#184](https://github.com/kubernetes/ingress/pull/184) Fix template error
- [X] [#179](https://github.com/kubernetes/ingress/pull/179) Allows the usage of Default SSL Cert
- [X] [#178](https://github.com/kubernetes/ingress/pull/178) Add initialization of proxy variable
- [X] [#177](https://github.com/kubernetes/ingress/pull/177) Refactoring sysctlFSFileMax helper
- [X] [#176](https://github.com/kubernetes/ingress/pull/176) Fix TLS does not get updated when changed
- [X] [#174](https://github.com/kubernetes/ingress/pull/174) Update nginx to 1.11.9
- [X] [#172](https://github.com/kubernetes/ingress/pull/172) add some unit test cases for some packages under folder "core.pkg.ingress"
- [X] [#168](https://github.com/kubernetes/ingress/pull/168) Changes the SSL Temp file to something inside the same SSL Directory
- [X] [#165](https://github.com/kubernetes/ingress/pull/165) Fix rate limit issue when more than 2 servers enabled in ingress
- [X] [#161](https://github.com/kubernetes/ingress/pull/161) Document some missing parameters and their defaults for NGINX controller
- [X] [#158](https://github.com/kubernetes/ingress/pull/158) prefect unit test cases for annotation.proxy
- [X] [#156](https://github.com/kubernetes/ingress/pull/156) Fix issue for ratelimit
- [X] [#154](https://github.com/kubernetes/ingress/pull/154) add unit test cases for core.pkg.ingress.annotations.cors
- [X] [#151](https://github.com/kubernetes/ingress/pull/151) Port in redirect
- [X] [#150](https://github.com/kubernetes/ingress/pull/150) Add support for custom header sizes
- [X] [#149](https://github.com/kubernetes/ingress/pull/149) Add flag to allow switch off the update of Ingress status
- [X] [#148](https://github.com/kubernetes/ingress/pull/148) Add annotation to allow custom body sizes
- [X] [#145](https://github.com/kubernetes/ingress/pull/145) fix wrong links and punctuations
- [X] [#144](https://github.com/kubernetes/ingress/pull/144) add unit test cases for core.pkg.k8s
- [X] [#143](https://github.com/kubernetes/ingress/pull/143) Use protobuf instead of rest to connect to apiserver host and add troubleshooting doc
- [X] [#142](https://github.com/kubernetes/ingress/pull/142) Use system fs.max-files as limits instead of hard-coded value
- [X] [#141](https://github.com/kubernetes/ingress/pull/141) Add reuse port and backlog to port 80 and 443
- [X] [#138](https://github.com/kubernetes/ingress/pull/138) reference to const
- [X] [#136](https://github.com/kubernetes/ingress/pull/136) Add content and descriptions about nginx's configuration
- [X] [#135](https://github.com/kubernetes/ingress/pull/135) correct improper punctuation
- [X] [#134](https://github.com/kubernetes/ingress/pull/134) fix typo
- [X] [#133](https://github.com/kubernetes/ingress/pull/133) Add TCP and UDP services removed in migration
- [X] [#132](https://github.com/kubernetes/ingress/pull/132) Document nginx controller configuration tweaks
- [X] [#128](https://github.com/kubernetes/ingress/pull/128) Add tests and godebug to compare structs
- [X] [#126](https://github.com/kubernetes/ingress/pull/126) change the type of imagePullPolicy
- [X] [#123](https://github.com/kubernetes/ingress/pull/123) Add resolver configuration to nginx
- [X] [#119](https://github.com/kubernetes/ingress/pull/119) add unit test case for annotations.service
- [X] [#115](https://github.com/kubernetes/ingress/pull/115) add default_server to listen statement for default backend
- [X] [#114](https://github.com/kubernetes/ingress/pull/114) fix typo
- [X] [#113](https://github.com/kubernetes/ingress/pull/113) Add condition of enqueue and unit test cases for task.Queue
- [X] [#108](https://github.com/kubernetes/ingress/pull/108) annotations: print error and skip if malformed
- [X] [#107](https://github.com/kubernetes/ingress/pull/107) fix some wrong links of examples which to be used for nginx
- [X] [#103](https://github.com/kubernetes/ingress/pull/103) Update the nginx controller manifests
- [X] [#101](https://github.com/kubernetes/ingress/pull/101) Add unit test for strings.StringInSlice
- [X] [#99](https://github.com/kubernetes/ingress/pull/99) Update nginx to 1.11.8
- [X] [#97](https://github.com/kubernetes/ingress/pull/97) Fix gofmt
- [X] [#96](https://github.com/kubernetes/ingress/pull/96) Fix typo PassthrougBackends -> PassthroughBackends
- [X] [#95](https://github.com/kubernetes/ingress/pull/95) Deny location mapping in case of specific errors
- [X] [#94](https://github.com/kubernetes/ingress/pull/94) Add support to disable server_tokens directive
- [X] [#93](https://github.com/kubernetes/ingress/pull/93) Fix sort for catch all server
- [X] [#92](https://github.com/kubernetes/ingress/pull/92) Refactoring of nginx configuration deserialization
- [X] [#91](https://github.com/kubernetes/ingress/pull/91) Fix x-forwarded-port mapping
- [X] [#90](https://github.com/kubernetes/ingress/pull/90) fix the wrong link to build/test/release
- [X] [#89](https://github.com/kubernetes/ingress/pull/89) fix the wrong links to the examples and developer documentation
- [X] [#88](https://github.com/kubernetes/ingress/pull/88) Fix multiple tls hosts sharing the same secretName
- [X] [#86](https://github.com/kubernetes/ingress/pull/86) Update X-Forwarded-Port
- [X] [#82](https://github.com/kubernetes/ingress/pull/82) Fix incorrect X-Forwarded-Port for TLS
- [X] [#81](https://github.com/kubernetes/ingress/pull/81) Do not push containers to remote repo as part of test-e2e
- [X] [#78](https://github.com/kubernetes/ingress/pull/78) Fix #76: hardcode X-Forwarded-Port due to SSL Passthrough
- [X] [#77](https://github.com/kubernetes/ingress/pull/77) Add support for IPV6 in dns resolvers
- [X] [#66](https://github.com/kubernetes/ingress/pull/66) Start FAQ docs
- [X] [#65](https://github.com/kubernetes/ingress/pull/65) Support hostnames in Ingress status
- [X] [#64](https://github.com/kubernetes/ingress/pull/64) Sort whitelist list to avoid random orders
- [X] [#62](https://github.com/kubernetes/ingress/pull/62) Fix e2e make targets
- [X] [#61](https://github.com/kubernetes/ingress/pull/61) Ignore coverage profile files
- [X] [#58](https://github.com/kubernetes/ingress/pull/58) Fix "invalid port in upstream" on nginx controller
- [X] [#57](https://github.com/kubernetes/ingress/pull/57) Fix invalid port in upstream
- [X] [#54](https://github.com/kubernetes/ingress/pull/54) Expand developer docs
- [X] [#52](https://github.com/kubernetes/ingress/pull/52) fix typo in variable ProxyRealIPCIDR
- [X] [#44](https://github.com/kubernetes/ingress/pull/44) Bump nginx version to one higher than that in contrib
- [X] [#36](https://github.com/kubernetes/ingress/pull/36) Add nginx metrics to prometheus
- [X] [#34](https://github.com/kubernetes/ingress/pull/34) nginx: also listen on ipv6
- [X] [#32](https://github.com/kubernetes/ingress/pull/32) Restart nginx if master process dies
- [X] [#31](https://github.com/kubernetes/ingress/pull/31) Add healthz checker
- [X] [#25](https://github.com/kubernetes/ingress/pull/25) Fix a data race in TestFileWatcher
- [X] [#12](https://github.com/kubernetes/ingress/pull/12) Split implementations from generic code
- [X] [#10](https://github.com/kubernetes/ingress/pull/10) Copy Ingress history from kubernetes/contrib
- [X] [#1498](https://github.com/kubernetes/contrib/pull/1498) Refactoring of template handling
- [X] [#1571](https://github.com/kubernetes/contrib/pull/1571) use POD_NAMESPACE as a namespace in cli parameters
- [X] [#1591](https://github.com/kubernetes/contrib/pull/1591) Always listen on port 443, even without ingress rules
- [X] [#1596](https://github.com/kubernetes/contrib/pull/1596) Adapt nginx hash sizes to the number of ingress
- [X] [#1653](https://github.com/kubernetes/contrib/pull/1653) Update image version
- [X] [#1672](https://github.com/kubernetes/contrib/pull/1672) Add firewall rules and ing class clarifications
- [X] [#1711](https://github.com/kubernetes/contrib/pull/1711) Add function helpers to nginx template
- [X] [#1743](https://github.com/kubernetes/contrib/pull/1743) Allow customisation of the nginx proxy_buffer_size directive via ConfigMap
- [X] [#1749](https://github.com/kubernetes/contrib/pull/1749) Readiness probe that works behind a CP lb
- [X] [#1751](https://github.com/kubernetes/contrib/pull/1751) Add the name of the upstream in the log
- [X] [#1758](https://github.com/kubernetes/contrib/pull/1758) Update nginx to 1.11.4
- [X] [#1759](https://github.com/kubernetes/contrib/pull/1759) Add support for default backend in Ingress rule
- [X] [#1762](https://github.com/kubernetes/contrib/pull/1762) Add cloud detection
- [X] [#1766](https://github.com/kubernetes/contrib/pull/1766) Clarify the controller uses endpoints and not services
- [X] [#1767](https://github.com/kubernetes/contrib/pull/1767) Update godeps
- [X] [#1772](https://github.com/kubernetes/contrib/pull/1772) Avoid replacing nginx.conf file if the new configuration is invalid
- [X] [#1773](https://github.com/kubernetes/contrib/pull/1773) Add annotation to add CORS support
- [X] [#1786](https://github.com/kubernetes/contrib/pull/1786) Add docs about go template
- [X] [#1796](https://github.com/kubernetes/contrib/pull/1796) Add external authentication support using auth_request
- [X] [#1802](https://github.com/kubernetes/contrib/pull/1802) Initialize proxy_upstream_name variable
- [X] [#1806](https://github.com/kubernetes/contrib/pull/1806) Add docs about the log format
- [X] [#1808](https://github.com/kubernetes/contrib/pull/1808) WebSocket documentation
- [X] [#1847](https://github.com/kubernetes/contrib/pull/1847) Change structure of packages
- [X] Add annotation for custom upstream timeouts
- [X] Mutual TLS auth (https://github.com/kubernetes/contrib/issues/1870)

### 0.8.3

- [X] [#1450](https://github.com/kubernetes/contrib/pull/1450) Check for errors in nginx template
- [ ] [#1498](https://github.com/kubernetes/contrib/pull/1498) Refactoring of template handling
- [X] [#1467](https://github.com/kubernetes/contrib/pull/1467) Use ClientConfig to configure connection
- [X] [#1575](https://github.com/kubernetes/contrib/pull/1575) Update nginx to 1.11.3

### 0.8.2

- [X] [#1336](https://github.com/kubernetes/contrib/pull/1336) Add annotation to skip ingress rule
- [X] [#1338](https://github.com/kubernetes/contrib/pull/1338) Add HTTPS default backend
- [X] [#1351](https://github.com/kubernetes/contrib/pull/1351) Avoid generation of invalid ssl certificates
- [X] [#1379](https://github.com/kubernetes/contrib/pull/1379) improve nginx performance
- [X] [#1350](https://github.com/kubernetes/contrib/pull/1350) Improve performance (listen backlog=net.core.somaxconn)
- [X] [#1384](https://github.com/kubernetes/contrib/pull/1384) Unset Authorization header when proxying
- [X] [#1398](https://github.com/kubernetes/contrib/pull/1398) Mitigate HTTPoxy Vulnerability

### 0.8.1

- [X] [#1317](https://github.com/kubernetes/contrib/pull/1317) Fix duplicated real_ip_header
- [X] [#1315](https://github.com/kubernetes/contrib/pull/1315) Addresses #1314

### 0.8

- [X] [#1063](https://github.com/kubernetes/contrib/pull/1063) watches referenced tls secrets
- [X] [#850](https://github.com/kubernetes/contrib/pull/850) adds configurable SSL redirect nginx controller
- [X] [#1136](https://github.com/kubernetes/contrib/pull/1136) Fix nginx rewrite rule order
- [X] [#1144](https://github.com/kubernetes/contrib/pull/1144) Add cidr whitelist support
- [X] [#1230](https://github.com/kubernetes/contrib/pull/1130) Improve docs and examples
- [X] [#1258](https://github.com/kubernetes/contrib/pull/1258) Avoid sync without a reachable
- [X] [#1235](https://github.com/kubernetes/contrib/pull/1235) Fix stats by country in nginx status page
- [X] [#1236](https://github.com/kubernetes/contrib/pull/1236) Update nginx to add dynamic TLS records and spdy
- [X] [#1238](https://github.com/kubernetes/contrib/pull/1238) Add support for dynamic TLS records and spdy
- [X] [#1239](https://github.com/kubernetes/contrib/pull/1239) Add support for conditional log of urls
- [X] [#1253](https://github.com/kubernetes/contrib/pull/1253) Use delayed queue
- [X] [#1296](https://github.com/kubernetes/contrib/pull/1296) Fix formatting
- [X] [#1299](https://github.com/kubernetes/contrib/pull/1299) Fix formatting

### 0.7

- [X] [#898](https://github.com/kubernetes/contrib/pull/898) reorder locations. Location / must be the last one to avoid errors routing to subroutes
- [X] [#946](https://github.com/kubernetes/contrib/pull/946) Add custom authentication (Basic or Digest) to ingress rules
- [X] [#926](https://github.com/kubernetes/contrib/pull/926) Custom errors should be optional
- [X] [#1002](https://github.com/kubernetes/contrib/pull/1002) Use k8s probes (disable NGINX checks)
- [X] [#962](https://github.com/kubernetes/contrib/pull/962) Make optional http2
- [X] [#1054](https://github.com/kubernetes/contrib/pull/1054) force reload if some certificate change
- [X] [#958](https://github.com/kubernetes/contrib/pull/958) update NGINX to 1.11.0 and add digest module
- [X] [#960](https://github.com/kubernetes/contrib/issues/960) https://trac.nginx.org/nginx/changeset/ce94f07d50826fcc8d48f046fe19d59329420fdb/nginx
- [X] [#1057](https://github.com/kubernetes/contrib/pull/1057) Remove loadBalancer ip on shutdown
- [X] [#1079](https://github.com/kubernetes/contrib/pull/1079) path rewrite
- [X] [#1093](https://github.com/kubernetes/contrib/pull/1093) rate limiting
- [X] [#1102](https://github.com/kubernetes/contrib/pull/1102) geolocation of traffic in stats
- [X] [#884](https://github.com/kubernetes/contrib/issues/884) support services running ssl
- [X] [#930](https://github.com/kubernetes/contrib/issues/930) detect changes in configuration configmaps
