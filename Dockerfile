ARG HELM_VERSION=3.2.1
FROM alpine/helm:${HELM_VERSION}

ADD release/linux/amd64/helm-push-plugin /bin/

ENTRYPOINT  ["/bin/helm-push-plugin"]
