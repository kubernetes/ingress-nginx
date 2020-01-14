FROM quay.io/kubernetes-ingress-controller/e2e:v01142020-3f0df1c35 AS BASE

FROM alpine:3.11

RUN apk add -U --no-cache \
    ca-certificates \
    bash \
    curl \
    tzdata \
    libc6-compat \
    openssl

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/
COPY --from=BASE /usr/local/bin/kubectl /usr/local/bin/

COPY e2e.sh             /e2e.sh
COPY cloud-generic      /cloud-generic
COPY cluster-wide       /cluster-wide
COPY overlay            /overlay
COPY namespace-overlays /namespace-overlays
RUN sed -E -i 's|^- .*deploy/cloud-generic$|- ../cloud-generic|' /overlay/kustomization.yaml
COPY wait-for-nginx.sh  /
COPY e2e.test           /

CMD [ "/e2e.sh" ]
