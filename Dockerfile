FROM golang:1.11.4-stretch
COPY . /go/src/deck
WORKDIR /go/src/deck
RUN go get .
RUN CGO_ENABLED=0 GOOS=linux go install -a -ldflags '-s -w -extldflags "-static"' .
ENTRYPOINT ["/bin/cp", "-v", "/go/bin/deck", "/out"]
