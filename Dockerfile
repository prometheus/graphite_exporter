FROM  quay.io/prometheus/busybox:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

COPY graphite_exporter /bin/graphite_exporter

USER        nobody
EXPOSE      9108 9109 9109/udp
ENTRYPOINT  [ "/bin/graphite_exporter" ]
