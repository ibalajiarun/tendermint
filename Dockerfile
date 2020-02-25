FROM alpine:3.7

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq file bash

RUN mkdir -p $HOME/.tendermint/config $HOME/.tendermint/data

# VOLUME [ /tendermint ]
WORKDIR /tendermint
EXPOSE 26656 26657
# ENTRYPOINT ["/usr/bin/wrapper.sh"]
# CMD ["node", "--proxy_app", "kvstore"]
STOPSIGNAL SIGTERM

COPY build/tendermint /tendermint/tendermint
COPY k8s/setup-server.sh /tendermint/setup-server.sh
# COPY wrapper.sh /usr/bin/wrapper.sh
# COPY config-template.toml /etc/tendermint/config-template.toml
