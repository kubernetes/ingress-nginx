FROM k8s.gcr.io/ingress-nginx/e2e-test-runner:v20210822-g5e5faa24d@sha256:55c568d9e35e15d94b3ab41fe549b8ee4cd910cc3e031ddcccd06256755c5d89 AS BASE

FROM alpine:3.12

RUN apk add -U --no-cache \
    ca-certificates \
    bash \
    curl \
    tzdata \
    libc6-compat \
    openssl

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/
COPY --from=BASE /usr/local/bin/helm /usr/local/bin/
COPY --from=BASE /usr/local/bin/kubectl /usr/local/bin/
COPY --from=BASE /usr/bin/cfssl /usr/local/bin/
COPY --from=BASE /usr/bin/cfssljson /usr/local/bin/

COPY . /

CMD [ "/e2e.sh" ]
