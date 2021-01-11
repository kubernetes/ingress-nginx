FROM squidfunk/mkdocs-material:6.2.4

RUN pip install mkdocs-awesome-pages-plugin

COPY action.sh /action.sh

RUN apk add --no-cache bash \
  && chmod +x /action.sh

ENTRYPOINT ["/action.sh"]
