ARG GO_VERSION=1.25.7@sha256:931c889bca758a82fcbfcb1b6ed6ca1de30783e9e52e6093ad50060735cb99be
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

FROM alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715
RUN adduser --disabled-password --gecos "" deckuser
RUN apk --no-cache add ca-certificates jq
USER deckuser
COPY --from=build /deck/deck /usr/local/bin
ENTRYPOINT ["deck"]
