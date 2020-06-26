ARG BASE_IMAGE

FROM ${BASE_IMAGE}

RUN apk add -U perl curl make unzip

ARG LUAROCKS_VERSION
ARG LUAROCKS_SHA

RUN wget -O /tmp/luarocks.tgz \
  https://github.com/luarocks/luarocks/archive/v${LUAROCKS_VERSION}.tar.gz \
  && echo "${LUAROCKS_SHA} */tmp/luarocks.tgz" | sha256sum -c - \
  && tar -C /tmp -xzf /tmp/luarocks.tgz \
  && cd /tmp/luarocks* \
  && ./configure \
  && make install

RUN luarocks install lua-resty-template

COPY nginx.conf /etc/nginx/nginx.conf
