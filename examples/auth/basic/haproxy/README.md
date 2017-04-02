# HAProxy Ingress Basic Authentication

This example demonstrates how to configure
[Basic Authentication](https://tools.ietf.org/html/rfc2617) on
HAProxy Ingress controller.

## Prerequisites

This document has the following prerequisites:

* Deploy [HAProxy Ingress controller](/examples/deployment/haproxy), you should
end up with controller, a sample web app and an ingress resource to the `foo.bar`
domain

As mentioned in the deployment instructions, you MUST turn down any existing
ingress controllers before running HAProxy Ingress.

## Using Basic Authentication

HAProxy Ingress read user and password from `auth` file stored on secrets, one user
and password per line. Secret name, realm and type are configured with annotations
in the ingress resource:

* `ingress.kubernetes.io/auth-type`: the only supported type is `basic`
* `ingress.kubernetes.io/auth-realm`: an optional string with authentication realm
* `ingress.kubernetes.io/auth-secret`: name of the secret

Each line of the `auth` file should have:

* user and insecure password separated with a pair of colons: `<username>::<plain-text-passwd>`; or
* user and an encrypted password separated with colons: `<username>:<encrypted-passwd>`

HAProxy evaluates encrypted passwords with
[crypt](http://man7.org/linux/man-pages/man3/crypt.3.html) function. Use `mkpasswd` or
`makepasswd` to create it. `mkpasswd` can be found on Alpine Linux container.

## Configure

Create a secret to our users:

* `john` and password `admin` using insecure plain text password
* `jane` and password `guest` using encrypted password

```console
$ mkpasswd -m des ## a short, des encryption, syntax from Busybox on Alpine Linux
Password: (type 'guest' and press Enter)
E5BrlrQ5IXYK2

$ cat >auth <<EOF
john::admin
jane:E5BrlrQ5IXYK2
EOF

$ kubectl create secret generic mypasswd --from-file auth
$ rm -fv auth
```

Annotate the ingress resource created on a [previous step](/examples/deployment/haproxy):

```console
$ kubectl annotate ingress/app \
    ingress.kubernetes.io/auth-type=basic \
    ingress.kubernetes.io/auth-realm="My Server" \
    ingress.kubernetes.io/auth-secret=mypasswd
```

Test without user and password:

```console
$ curl -i 172.17.4.99:30876 -H 'Host: foo.bar'
HTTP/1.0 401 Unauthorized
Cache-Control: no-cache
Connection: close
Content-Type: text/html
WWW-Authenticate: Basic realm="My Server"

<html><body><h1>401 Unauthorized</h1>
You need a valid user and password to access this content.
</body></html>
```

Send a valid user:

```console
$ curl -i -u 'john:admin' 172.17.4.99:30876 -H 'Host: foo.bar'
HTTP/1.1 200 OK
Server: nginx/1.9.11
Date: Sun, 05 Mar 2017 19:22:33 GMT
Content-Type: text/plain
Transfer-Encoding: chunked

CLIENT VALUES:
client_address=10.2.18.5
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://foo.bar:8080/
```

Using `jane:guest` user/passwd should have the same output.

