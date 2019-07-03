ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/graphite_exporter /bin/graphite_exporter

USER        nobody
EXPOSE      9108 9109 9109/udp
ENTRYPOINT  [ "/bin/graphite_exporter" ]
