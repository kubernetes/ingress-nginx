NGINX base image

### HTTP/3 Support

**HTTP/3 support is experimental and under development**

[HTTP/3](https://datatracker.ietf.org/doc/html/rfc9114)\
[QUIC](https://datatracker.ietf.org/doc/html/rfc9000)

[According to the documentation, NGINX 1.25.0 or higher supports HTTP/3:](https://nginx.org/en/docs/quic.html)

> Support for QUIC and HTTP/3 protocols is available since 1.25.0.

But this requires adding a new flag during the build:

> When configuring nginx, it is possible to enable QUIC and HTTP/3 using the --with-http_v3_module configuration parameter.

[We have added this flag](https://github.com/kubernetes/ingress-nginx/pull/11470), but it is not enough to use HTTP/3 in ingress-nginx, this is the first step.

The next steps will be:

1. **Waiting for OpenSSL 3.4.**\
    The main problem is, that we still use OpenSSL (3.x) and it does not support the important mechanism of TLS 1.3 - [early_data](https://datatracker.ietf.org/doc/html/rfc8446#section-2.3):

    > Otherwise, the OpenSSL compatibility layer will be used that does not support early data.
    
    [And although another part of the documentation says that the directive is supported with OpenSSL:](https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_early_data)

    > The directive is supported when using OpenSSL 1.1.1 or higher.
    
    But this is incomplete support, because OpenSSL does not support this feature, and [it has only client side support:](https://github.com/openssl/openssl)

    > ... the QUIC (currently client side only) version 1 protocol
   
    [And also there are some issues even with client side](https://github.com/openssl/openssl/discussions/23339)

    Due to this, we currently have incomplete HTTP/3 support, without important security and performance features.\
    But the good news is that [OpenSSL plans to add server-side support in 3.4](https://github.com/openssl/web/blob/master/roadmap.md):

    > Server-side QUIC support
    
    [Overview of SSL libraries(HAProxy Documentation)](https://github.com/haproxy/wiki/wiki/SSL-Libraries-Support-Status#tldr)  

2. **Adding [parameters](https://nginx.org/en/docs/http/ngx_http_v3_module.html) to the configmap to configure HTTP/3 and quic(enableHTTP3, enableHTTP/0.9, maxCurrentStream, and so on).**
3. **Adding options to the nginx config template(`listen 443 quic` to server blocks and `add_header Alt-Svc 'h3=":8443"; ma=86400';` to location blocks).**
4. **Opening the https port for UDP in the container(because QUIC uses UDP).**
5. **Adding tests.**
