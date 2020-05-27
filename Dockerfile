FROM alpine as downloader

ARG HELM_VERSION=3.2.1
ENV HELM_URL=https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz

WORKDIR /tmp
RUN true \
  && wget -O helm.tgz "$HELM_URL" \
  && tar xvpf helm.tgz linux-amd64/helm \
  && mv linux-amd64/helm /usr/local/bin/helm

# ---

FROM busybox:glibc

COPY --from=downloader /usr/local/bin/helm /usr/local/bin/helm

ADD release/linux/amd64/helm-push-plugin /usr/local/bin/

ENTRYPOINT  ["/usr/local/bin/helm-push-plugin"]
