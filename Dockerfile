FROM golang:1.13.2 AS build
WORKDIR /deck
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o deck

FROM alpine:3.10
RUN adduser --disabled-password --gecos "" deckuser
RUN apk --no-cache add ca-certificates
USER deckuser
COPY --from=build /deck/deck /usr/local/bin
ENTRYPOINT ["deck"]
