FROM golang:1.27.1-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 as builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v12.3.4@sha256:8359cbcf2c16ca7dcafd217e9afaeae66a0fc752099c868221cc22d2587ea60b . /

COPY . /src/share
WORKDIR /src/share

RUN set -ex \
 && apk add --update \
      nodejs \
      npm \
      git \
      make \
 && make frontend \
 && go install \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
      -mod=readonly \
      -modcacherw \
      -trimpath


FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

LABEL maintainer="Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add \
      ca-certificates

COPY --from=builder /go/bin/share /usr/local/bin/share

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/share"]
CMD ["--"]

# vim: set ft=Dockerfile:
