# Changelog

## 1.2.1-0.5.0 (upcoming)

* Build the stratio ingress-nginx-controller with the community controller:1.2.1 as a base

## 1.2.0-0.4.0 (2022-05-11)

* Build the stratio ingress-nginx-controller with the community controller:1.2.0 as a base

## Previous development

### Branched to branch-0.3 (2022-04-07)

* Fix openssl CVE-2022-0778
* [EOS-6280] Move vault annotations to annotations with the prefix 'nginx.ingress.stratio.com' 
* [EOS-5973] Use certificates from vault by reading annotations
* [EOS-6019] Accept flag --default-ssl-certificate-vault for reading the default certificate from vault secrets

### Branched to branch-0.2 (2022-02-25)

* [EOS-5623] Use RSA keys for jwt signing
* [EOS-6014] Setting SameSite property of stratio session cookie to 'lax'  

### Branched to branch-0.1 (2021-09-02)

* [EOS-5356] Fix: return last CA for authn purposes
* [EOS-5356] Return full CA chain
* Fix: Allow no http_cookie for requests with certs
* Fix: Add path / to stratio-cookie
* Add cert authentication
* Add Vault integration
* Adapt repo to Stratio CICD flow 
