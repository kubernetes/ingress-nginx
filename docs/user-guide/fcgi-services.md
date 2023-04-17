

# Exposing FastCGI Servers

**This feature has been removed from Ingress NGINX**

People willing to use fastcgi servers, should create an NGINX + FastCGI service and expose
this service via Ingress NGINX. 

We recommend using images like `cgr.dev/chainguard/nginx:latest` and expose your fast_cgi application
as another container on this Pod.
