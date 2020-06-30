FROM us.gcr.io/k8s-artifacts-prod/ingress-nginx/e2e-test-runner@sha256:7dece116c5cc1a51496095de97cc9249a07e9f1a4ed0dc378630074c6857ff46 AS BASE

FROM alpine:3.12

RUN apk add -U --no-cache \
    ca-certificates \
    bash \
    curl \
    tzdata \
    libc6-compat \
    openssl

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/
COPY --from=BASE /usr/local/bin/kubectl /usr/local/bin/
COPY --from=BASE /usr/local/bin/cfssl /usr/local/bin/
COPY --from=BASE /usr/local/bin/helm /usr/local/bin/
COPY --from=BASE /usr/local/bin/cfssljson /usr/local/bin/

COPY . /

CMD [ "/e2e.sh" ]
