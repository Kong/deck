FROM golang:1.22.4 AS build
WORKDIR /deck
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD . .
ARG COMMIT
ARG TAG
RUN CGO_ENABLED=0 GOOS=linux go build -o deck \
      -ldflags "-s -w -X github.com/kong/deck/cmd.VERSION=$TAG -X github.com/kong/deck/cmd.COMMIT=$COMMIT"

FROM alpine:3.20.2
RUN adduser --disabled-password --gecos "" deckuser
RUN apk --no-cache add ca-certificates jq
USER deckuser
COPY --from=build /deck/deck /usr/local/bin
ENTRYPOINT ["deck"]
