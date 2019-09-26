ARG IMAGE_BASE=alpine:3.10

FROM golang:1.13.0 AS build
WORKDIR /deck
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o deck

FROM ${IMAGE_BASE}
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /deck/deck /usr/local/bin
COPY ./docker_entrypoint.sh /docker_entrypoint.sh
ENTRYPOINT ["/docker_entrypoint.sh"]
