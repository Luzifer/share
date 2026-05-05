FROM golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 as builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v10.33.2@sha256:11b50efc9a57a215c6e33f6d20f27cf42004d06919a451bffae86a314db6fee8 . /

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
