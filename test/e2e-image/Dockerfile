FROM k8s.gcr.io/ingress-nginx/e2e-test-runner:v20210810-g820a21a74@sha256:7d7393a8c6c72d76145282df53ea0679a5b769211fd1cd6b8910b6dda1bd986d AS BASE

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
