FROM registry.k8s.io/ingress-nginx/e2e-test-runner:v20220624-g3348cd71e@sha256:2a34e322b7ff89abdfa0b6202f903bf5618578b699ff609a3ddabac0aae239c8 AS BASE

FROM alpine:3.14.6

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
