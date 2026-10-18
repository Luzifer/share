FROM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d as builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v11.28.6@sha256:d35ec722ba2d0bc7d8ff9d9920d2b9f851e566d13822e33da1272d085a5b03a8 . /

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
