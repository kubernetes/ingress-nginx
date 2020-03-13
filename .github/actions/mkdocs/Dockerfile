FROM squidfunk/mkdocs-material:4.6.3

COPY action.sh /action.sh

RUN apk add --no-cache bash \
  && chmod +x /action.sh

ENTRYPOINT ["/action.sh"]
