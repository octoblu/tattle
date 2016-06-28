FROM golang:1.6
MAINTAINER Octoblu, Inc. <docker@octoblu.com>

WORKDIR /go/src/github.com/octoblu/tattle
COPY . /go/src/github.com/octoblu/tattle

RUN env CGO_ENABLED=0 go build -o tattle -a -ldflags '-s' .

CMD ["./tattle"]
