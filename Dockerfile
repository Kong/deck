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
      -ldflags "-s -w -X github.com/kong/deck/cmd.VERSION=$TAG -X github.com/kong/deck/cmd.COMMIT=$COMMIT"

FROM ghcr.io/jqlang/jq:1.8.2@sha256:b9c68867e5766576263a222e91db3de422d802069c7af70440e667a95344e486 AS jq

FROM gcr.io/distroless/base:latest@sha256:f4a335ca209e1d2ee873102c17c389ad0142e3d5b21aee2817e9cc9c01d87d20

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
USER nonroot
COPY --from=build /deck/deck /usr/local/bin/
COPY --from=jq /jq /usr/local/bin/jq
ENTRYPOINT ["deck"]