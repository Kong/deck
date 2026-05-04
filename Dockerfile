ARG GO_VERSION=1.25.9@sha256:8a7adc288b77e9b787cd2695029eb54d10ae80571b21d44fed68d067ad0a9c96
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

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
RUN adduser --disabled-password --gecos "" deckuser
RUN apk --no-cache upgrade && apk --no-cache add ca-certificates jq
USER deckuser
COPY --from=build /deck/deck /usr/local/bin
ENTRYPOINT ["deck"]