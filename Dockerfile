ARG HELM_VERSION=3.2.1
FROM alpine/helm:${HELM_VERSION}

RUN apk --update add git less openssh && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

RUN helm plugin install https://github.com/chartmuseum/helm-push

ADD release/linux/amd64/helm-push-plugin /bin/

ENTRYPOINT  ["/bin/helm-push-plugin"]
