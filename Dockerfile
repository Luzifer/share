FROM golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 as builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v11.0.3@sha256:61d56059765bcdc7480753f8b095e2467d799cf9103244521a2f7aac088417ff . /

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


FROM alpine:3.23@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

LABEL maintainer="Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add \
      ca-certificates

COPY --from=builder /go/bin/share /usr/local/bin/share

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/share"]
CMD ["--"]

# vim: set ft=Dockerfile:
