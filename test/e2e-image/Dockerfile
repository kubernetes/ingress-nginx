ARG E2E_BASE_IMAGE
FROM ${E2E_BASE_IMAGE} AS BASE

FROM alpine:3.17.2

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
