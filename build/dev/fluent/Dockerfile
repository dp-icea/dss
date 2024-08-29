FROM fluent/fluentd:v1.17
USER root

RUN apk add --no-cache --update --virtual .build-deps \
        sudo build-base ruby-dev \
 && gem install fluent-plugin-opensearch \
 && gem sources --clear-all \
 && apk del .build-deps \
 && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem

EXPOSE 24224
EXPOSE 9880 