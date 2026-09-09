ARG GO_VERSION=1.26.6@sha256:0d1d3a794be25f809dd2cb3160d8c73276c4056a9f8242a138e908ddeee7b6b6
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

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6

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
