ARG E2E_BASE_IMAGE
FROM ${E2E_BASE_IMAGE} AS BASE

FROM alpine:3.23.2

RUN apk update \
	&& apk upgrade && apk add -U --no-cache \
  ca-certificates \
  bash \
  tzdata \
  libc6-compat \
  openssl \
  sqlite-dev

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/
COPY --from=BASE /usr/local/bin/helm /usr/local/bin/
COPY --from=BASE /usr/local/bin/kubectl /usr/local/bin/
COPY --from=BASE /usr/bin/cfssl /usr/local/bin/
COPY --from=BASE /usr/bin/cfssljson /usr/local/bin/

COPY . /

CMD [ "/e2e.sh" ]
