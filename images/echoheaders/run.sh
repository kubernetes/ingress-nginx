#!/bin/bash

echo "Generating self-signed cert"
mkdir -p /certs
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 \
-keyout /certs/privateKey.key \
-out /certs/certificate.crt \
-subj "/C=UK/ST=Warwickshire/L=Leamington/O=OrgName/OU=IT Department/CN=example.com"

echo "Starting nginx"
nginx -g "daemon off;"