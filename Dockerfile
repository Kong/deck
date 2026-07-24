ARG GO_VERSION=1.26.4@sha256:f96cc555eb8db430159a3aa6797cd5bae561945b7b0fe7d0e284c63a3b291609
FROM golang:${GO_VERSION} AS build
WORKDIR /deck
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD . .
ARG COMMIT
ARG TAG
RUN CGO_ENABLED=0 GOOS=linux go build -o deck \
      -ldflags "-s -w -X github.com/kong/deck/cmd.VERSION=$TAG -X github.com/kong/deck/cmd.COMMIT=$COMMIT" -buildvcs=true

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
ARG COMMIT
ARG TAG
LABEL org.opencontainers.image.title="deck" \
      org.opencontainers.image.description="Declarative configuration for Kong" \
      org.opencontainers.image.url="https://github.com/kong/deck" \
      org.opencontainers.image.source="https://github.com/kong/deck" \
      org.opencontainers.image.version="$TAG" \
      org.opencontainers.image.revision="$COMMIT" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.vendor="Kong Inc."
RUN adduser --disabled-password --gecos "" deckuser
RUN apk --no-cache upgrade && apk --no-cache add ca-certificates jq
USER deckuser
COPY --from=build /deck/deck /usr/local/bin
ENTRYPOINT ["deck"]