FROM k8s.gcr.io/ingress-nginx/e2e-test-runner:v20210906-g7d577d976@sha256:cf7079b5c05b8b1b108b16752c6ff4ca312cf96700e91eef6088b9e0c4a7aff1 AS BASE

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
