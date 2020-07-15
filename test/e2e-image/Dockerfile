FROM us.gcr.io/k8s-artifacts-prod/ingress-nginx/e2e-test-runner@sha256:bf31d6f13984ef0818312f9247780e234a6d50322e362217dabe4b3299f202ef AS BASE

FROM alpine:3.12

RUN apk add -U --no-cache \
    ca-certificates \
    bash \
    curl \
    tzdata \
    libc6-compat \
    openssl

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/
COPY --from=BASE /usr/local/bin/kubectl /usr/local/bin/
COPY --from=BASE /usr/local/bin/cfssl /usr/local/bin/
COPY --from=BASE /usr/local/bin/helm /usr/local/bin/
COPY --from=BASE /usr/local/bin/cfssljson /usr/local/bin/

COPY . /

CMD [ "/e2e.sh" ]
