FROM k8s.gcr.io/ingress-nginx/e2e-test-runner:v20200904-gc2884a3da@sha256:f2de204dafb3951b974a53f7a4e9cc43b62b93c99b95edcb7a65afdc1d4f0bf4 AS BASE

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
