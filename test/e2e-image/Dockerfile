FROM k8s.gcr.io/ingress-nginx/e2e-test-runner:v20220416-controller-v1.2.0-beta.0-4-g2e1a4790b@sha256:4468eb8cc9aa0ec3971ddf3811efe363e6f8e9082e95b567a5c27d91f89fb1e3   AS BASE

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
