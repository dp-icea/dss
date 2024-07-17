FROM fluent/fluentd:v1.16-1
USER root

RUN apk add --no-cache --update --virtual .build-deps \
        sudo build-base ruby-dev \
 && sudo gem install fluent-plugin-elasticsearch \
 && sudo gem sources --clear-all \
 && apk del .build-deps \
 && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem
RUN gem install fluent-plugin-opensearch

RUN mkdir -p $(pwd)/output && chmod 777 $(pwd)/output

EXPOSE 24224
EXPOSE 9880 