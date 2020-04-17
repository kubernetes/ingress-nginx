FROM quay.io/kubernetes-ingress-controller/e2e:v04142020-0257068b9 AS BASE

FROM alpine:3.11

RUN apk add -U --no-cache \
    ca-certificates \
    bash \
    curl \
    tzdata \
    libc6-compat \
    openssl

RUN curl -sSL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash \
  && helm repo add stable https://kubernetes-charts.storage.googleapis.com \
  && helm repo update

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/
COPY --from=BASE /usr/local/bin/kubectl /usr/local/bin/

COPY . /

CMD [ "/e2e.sh" ]
