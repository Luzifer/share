FROM golang:1.26-alpine@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d as builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v11.0.6@sha256:107603bc6200228d9743f31eef4e3e3b3b2c9ce198ce34b74be4076ab3ce038c . /

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
