FROM quay.io/kubernetes-ingress-controller/debian-base-amd64:0.1

RUN clean-install \
    ca-certificates \
    bash \
    tzdata

COPY ginkgo     /usr/local/bin/
COPY kubectl    /usr/local/bin/
COPY e2e.sh     /e2e.sh

COPY manifests  /manifests

COPY wait-for-nginx.sh /
COPY e2e.test   /

CMD [ "/e2e.sh" ]
