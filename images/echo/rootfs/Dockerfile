FROM quay.io/kubernetes-ingress-controller/nginx:e3c49c52f4b74fe47ad65d6f3266a02e8b6b622f

RUN apk add -U perl curl make unzip

RUN wget -O /tmp/luarocks.tgz https://github.com/luarocks/luarocks/archive/v3.3.1.tar.gz \
  && tar -C /tmp -xzf /tmp/luarocks.tgz \
  && cd /tmp/luarocks* \
  && ./configure \
  && make install \
  && rm -rf /tmp/*

RUN luarocks install lua-resty-template

COPY nginx.conf /etc/nginx/nginx.conf
