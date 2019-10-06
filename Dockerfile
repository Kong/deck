FROM golang:1.13.0 AS build
WORKDIR /deck
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o deck

FROM alpine:3.10
RUN apk update \
    && apk upgrade \
    && apk --no-cache add ca-certificates

ENV USER=appuser
ENV GROUP=deck
ENV UID=9999
ENV GID=9999

RUN addgroup --gid "$GID" "$GROUP" \
    && adduser \
    --disabled-password \
    --gecos "" \
    --ingroup "$GROUP" \
    --no-create-home \
    --uid "$UID" \
    "$USER"

COPY --from=build /deck/deck /usr/local/bin

USER $USER

ENTRYPOINT ["deck"]
CMD ["help"]
