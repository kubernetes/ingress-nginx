# Ingress examples

This directory contains a catalog of examples on how to run, configure and scale Ingress.  
Please review the [prerequisites](PREREQUISITES.md) before trying them.

Category | Name | Description | Complexity Level
---------| ---- | ----------- | ----------------
Apps | [Docker Registry](docker-registry/README.md) | TODO | TODO
Auth | [Basic authentication](auth/basic/README.md) | password protect your website | Intermediate
Auth | [Client certificate authentication](auth/client-certs/README.md) | secure your website with client certificate authentication | Intermediate
Auth | [External authentication plugin](auth/external-auth/README.md) | defer to an external authentication service | Intermediate
Auth | [OAuth external auth](auth/oauth-external-auth/README.md) | TODO | TODO
Customization | [Configuration snippets](customization/configuration-snippets/README.md) | customize nginx location configuration using annotations | Advanced
Customization | [Custom configuration](customization/custom-configuration/README.md) | TODO | TODO
Customization | [Custom DH parameters for perfect forward secrecy](customization/ssl-dh-param/README.md) | TODO | TODO
Customization | [Custom errors](customization/custom-errors/README.md) | serve custom error pages from the default backend | Intermediate
Customization | [Custom headers](customization/custom-headers/README.md) | set custom headers before sending traffic to backends  | Advanced
Customization | [External authentication with response header propagation](customization/external-auth-headers/README.md) | TODO | TODO
Customization | [Sysctl tuning](customization/sysctl/README.md) | TODO | TODO
Features | [Rewrite](rewrite/README.md) | TODO | TODO
Features | [Session stickiness](affinity/cookie/README.md) | route requests consistently to the same endpoint | Advanced
Scaling | [Static IP](static-ip/README.md) | a single ingress gets a single static IP |  Intermediate
TLS | [Multi TLS certificate termination](multi-tls/README.md) | TODO | TODO
TLS | [TLS termination](tls-termination/README.md) | TODO | TODO
