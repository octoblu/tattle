FROM golang
MAINTAINER Octoblu, Inc. <docker@octoblu.com>

ADD https://raw.githubusercontent.com/pote/gpm/v1.3.2/bin/gpm /go/bin/
RUN chmod +x /go/bin/gpm

COPY Godeps /go/src/github.com/octoblu/tattle/
WORKDIR /go/src/github.com/octoblu/tattle
RUN gpm install

COPY . /go/src/github.com/octoblu/tattle

RUN env CGO_ENABLED=0 go build -a -ldflags '-s' .

CMD ["./tattle"]
