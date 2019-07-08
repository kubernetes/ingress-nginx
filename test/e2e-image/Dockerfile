FROM quay.io/kubernetes-ingress-controller/e2e:v06262019-ecce3fd7b AS BASE

FROM quay.io/kubernetes-ingress-controller/debian-base-amd64:0.1

RUN clean-install \
    ca-certificates \
    bash \
    curl \
    tzdata

RUN curl -Lo /usr/local/bin/kubectl \
    https://storage.googleapis.com/kubernetes-release/release/v1.15.0/bin/linux/amd64/kubectl \
    && chmod +x /usr/local/bin/kubectl

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/

COPY e2e.sh             /e2e.sh
COPY cloud-generic      /cloud-generic
COPY cluster-wide       /cluster-wide
COPY overlay            /overlay
RUN sed -E -i 's|^- .*deploy/cloud-generic$|- ../cloud-generic|' /overlay/kustomization.yaml
COPY wait-for-nginx.sh  /
COPY e2e.test           /

CMD [ "/e2e.sh" ]
