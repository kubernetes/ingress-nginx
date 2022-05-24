FROM k8s.gcr.io/ingress-nginx/e2e-test-runner:v20220524-g8963ed17e@sha256:4fbcbeebd4c24587699b027ad0f0aa7cd9d76b58177a3b50c228bae8141bcf95 AS BASE

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
