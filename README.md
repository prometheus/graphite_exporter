# Graphite Exporter [![Build Status](https://travis-ci.org/prometheus/graphite_exporter.svg)][travis]

[![CircleCI](https://circleci.com/gh/prometheus/graphite_exporter/tree/master.svg?style=shield)][circleci]
[![Docker Repository on Quay](https://quay.io/repository/prometheus/graphite-exporter/status)][quay]
[![Docker Pulls](https://img.shields.io/docker/pulls/prom/graphite-exporter.svg?maxAge=604800)][hub]

An exporter for metrics exported in the [Graphite plaintext
protocol](http://graphite.readthedocs.org/en/latest/feeding-carbon.html#the-plaintext-protocol).
It accepts data over both TCP and UDP, and transforms and exposes them for
consumption by Prometheus.

This exporter is useful for exporting metrics from existing Graphite setups, as
well as for metrics which are not covered by the core Prometheus exporters such
as the [Node Exporter](https://github.com/prometheus/node_exporter).

## Usage

```
make
./graphite_exporter
```

Configure existing monitoring to send Graphite plaintext data to port 9109 on UDP or TCP.
As a simple demonstration:
```
echo "test_tcp 1234 $(date +%s)" | nc localhost 9109
echo "test_udp 1234 $(date +%s)" | nc -u -w1 localhost 9109
```

Metrics will be available on [http://localhost:9108/metrics](http://localhost:9108/metric).

To avoid using unbounded memory, metrics will be garbage collected five minutes after
they are last pushed to. This is configurable with the `-graphite.sample-expiry` flag.

## Metric Mapping and Configuration

The graphite_exporter can be configured to translate specific dot-separated
graphite metrics into labeled Prometheus metrics via a simple mapping language.
This is read from a file specified by the `-graphite.mapping-config` flag. A
mapping definition starts with a line matching the graphite metric in question,
with `*`s acting as wildcards for each dot-separated metric component. The
lines following the matching expression must contain one `label="value"` pair
each, and at least define the metric name (label name `name`). The Prometheus
metric is then constructed from these labels. `$n`-style references in the
label value are replaced by the n-th wildcard match in the matching line,
starting at 1. Multiple matching definitions are separated by one or more empty
lines. The first mapping rule that matches a graohite metric wins.

Metrics that don't match any mapping in the configuration file are translated
into Prometheus metrics without any labels and with `.` characters replaced with `_`.

If you have a very large set of metrics you may want to skip the ones that don't
match the mapping configuration. If that is the case you can force this behaviour
using the `-graphite.mapping-strict-match` flag, and it will only store those metrics
you really want.

An example mapping configuration:

    test.dispatcher.*.*.*
    name="dispatcher_events_total"
    processor="$1"
    action="$2"
    outcome="$3"
    job="test_dispatcher"

    *.signup.*.*
    name="signup_events_total"
    provider="$2"
    outcome="$3"
    job="${1}_server"

This would transform these example graphite metrics into Prometheus metrics as
follows:

    test.dispatcher.FooProcessor.send.success
     => dispatcher_events_total{processor="FooProcessor", action="send", outcome="success", job="test_dispatcher"}

    foo_product.signup.facebook.failure
     => signup_events_total{provider="facebook", outcome="failure", job="foo_product_server"}

    test.web-server.foo.bar
     => test_web__server_foo_bar{}

## Using Docker

You can deploy this exporter using the [prom/graphite-exporter][hub] Docker image.

For example:

```bash
docker pull prom/graphite-exporter

docker run -d -p 9108:9108 -p 9109:9109 -p 9109/udp:9109/udp
        -v $PWD/graphite_mapping.conf:/tmp/graphite_mapping.conf \
        prom/graphite-exporter -graphite.mapping-config=/tmp/graphite_mapping.conf
```


[circleci]: https://circleci.com/gh/prometheus/graphite_exporter
[hub]: https://hub.docker.com/r/prom/graphite-exporter/
[travis]: https://travis-ci.org/prometheus/graphite_exporter
[quay]: https://quay.io/repository/prometheus/graphite-exporter
