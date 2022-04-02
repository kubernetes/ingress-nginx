FROM k8s.gcr.io/ingress-nginx/e2e-test-runner:v20220331-controller-v1.1.2-31-gf1cb2b73c@sha256:baa326f5c726d65be828852943a259c1f0572883590b9081b7e8fa982d64d96e AS BASE

FROM alpine:3.14.4

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
