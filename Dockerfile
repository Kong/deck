ARG GO_VERSION=1.26.4@sha256:84e4dd2658fb85e1debf9fbe653aa386ed5bf108429ea641ab0a7db31a8ed5cc
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

FROM alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11
RUN adduser --disabled-password --gecos "" deckuser
RUN apk --no-cache upgrade && apk --no-cache add ca-certificates jq
USER deckuser
COPY --from=build /deck/deck /usr/local/bin
ENTRYPOINT ["deck"]