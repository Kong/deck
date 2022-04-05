FROM golang:1.18.0 AS build
WORKDIR /deck
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD . .
ARG COMMIT
ARG TAG
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o deck \
      -ldflags "-s -w -X github.com/kong/deck/cmd.VERSION=$TAG -X github.com/kong/deck/cmd.COMMIT=$COMMIT"

FROM alpine:3.15.4
RUN adduser --disabled-password --gecos "" deckuser
RUN apk --no-cache add ca-certificates
USER deckuser
COPY --from=build /deck/deck /usr/local/bin
ENTRYPOINT ["deck"]
