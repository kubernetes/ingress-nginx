FROM k8s.gcr.io/ingress-nginx/e2e-test-runner:v20220110-gfd820db46@sha256:273f7d9b1b2297cd96b4d51600e45d932186a1cc79d00d179dfb43654112fe8f AS BASE

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
