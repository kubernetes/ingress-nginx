# Docker images

Directory | Purpose
------------ | -------------
custom-error-pages | Example of Custom error pages for the NGINX Ingress controller
e2e | Image to run e2e tests
e2e-prow | Image to launch Prow jobs
fastcgi-helloserver | FastCGI application for e2e tests
grpc-fortune-teller | grpc server application for the nginx-ingress grpc example
httpbin | A simple HTTP Request & Response Service for e2e tests
mkdocs | Image to build the static documentation
nginx | NGINX base image using [alpine linux](https://www.alpinelinux.org)

:bangbang: Only the nginx image is meant to be published. The others are used as examples for some feature of the ingress controller or to run e2e tests.
