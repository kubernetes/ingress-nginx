# TLS termination

Before continue, follow [deploying HAProxy Ingress](/examples/deployment/haproxy) in order to have a functional ingress controller.

Update ingress resource in order to add tls termination to host `foo.bar`:

    kubectl replace -f ingress-tls-default.yaml

Trying default backend:

    curl -iL 172.17.4.99:30876            
    HTTP/1.1 404 Not Found
    Date: Tue, 07 Feb 2017 00:06:07 GMT
    Content-Length: 21
    Content-Type: text/plain; charset=utf-8

    default backend - 404

Now telling the controller we are `foo.bar`:

    curl -iL 172.17.4.99:30876 -H 'Host: foo.bar'
    HTTP/1.1 302 Found
    Cache-Control: no-cache
    Content-length: 0
    Location: https://foo.bar/
    Connection: close
    ^C

Note the `Location` header - this would redirect us to the correct server.

Checking the default certificate - change below `31692` to the TLS port:

    openssl s_client -connect 172.17.4.99:31692
    ...
    subject=/CN=localhost
    issuer=/CN=localhost
    ---

... and `foo.bar` certificate:

    openssl s_client -connect 172.17.4.99:31692 -servername foo.bar
    ...
    subject=/CN=localhost
    issuer=/CN=localhost
    ---

Let's create a new certificate to our domain:

    openssl req \
      -x509 -newkey rsa:2048 -nodes -days 365 \
      -keyout tls.key -out tls.crt -subj '/CN=foo.bar'
    kubectl create secret tls foobar-ssl --cert=tls.crt --key=tls.key
    rm -v tls.crt tls.key

... and reference in the ingress resource:

    kubectl replace -f ingress-tls-foobar.yaml 

Now `foo.bar` certificate should be used to terminate tls:

    openssl s_client -connect 172.17.4.99:31692
    ...
    subject=/CN=localhost
    issuer=/CN=localhost
    ---

    openssl s_client -connect 172.17.4.99:31692 -servername foo.bar
    ...
    subject=/CN=foo.bar
    issuer=/CN=foo.bar
    ---
