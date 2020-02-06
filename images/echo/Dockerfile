FROM openresty/openresty:1.15.8.2-alpine

RUN apk add -U perl curl \
  && opm get bungle/lua-resty-template

COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
