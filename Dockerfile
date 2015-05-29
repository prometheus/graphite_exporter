FROM alpine:3.1
MAINTAINER The Prometheus Authors <prometheus-developers@googlegroups.com>

ENV GOPATH /go
COPY . /go/src/github.com/prometheus/graphite_exporter

RUN apk add --update -t build-deps go git mercurial \
    && apk add -u musl && rm -rf /var/cache/apk/* \
    && go get github.com/tools/godep \
    && cd /go/src/github.com/prometheus/graphite_exporter \
    && go get -d && go build -o /bin/graphite_exporter \
    && rm -rf /go && apk del --purge build-deps

EXPOSE     9108 9109
ENTRYPOINT [ "/bin/graphite_exporter" ]
