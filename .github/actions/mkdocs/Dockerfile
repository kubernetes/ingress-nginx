FROM squidfunk/mkdocs-material:5.2.3

COPY action.sh /action.sh

RUN apk add --no-cache bash \
  && chmod +x /action.sh

ENTRYPOINT ["/action.sh"]
