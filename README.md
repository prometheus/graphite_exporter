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

Metrics will be available on [http://localhost:9108/metrics](http://localhost:9108/metrics).

To avoid using unbounded memory, metrics will be garbage collected five minutes after
they are last pushed to. This is configurable with the `--graphite.sample-expiry` flag.

## Metric Mapping and Configuration

**Please note there has been a breaking change in configuration after version 0.2.0.  The YAML style config from [statsd_exporter](https://github.com/prometheus/statsd_exporter) is now used.  See conversion instructions below**

### YAML Config
The graphite_exporter can be configured to translate specific dot-separated
graphite metrics into labeled Prometheus metrics via YAML configuration file.  This file shares syntax and logic with [statsd_exporter](https://github.com/prometheus/statsd_exporter).  Please follow the statsd_exporter documentation for usage information.  However, graphite_exporter does not support *all* parsing features at this time.  Any feature based on the 'timer_type' option will not function.  Otherwise, regex matching, groups, match/drop behavior, should work as expected.

Metrics that don't match any mapping in the configuration file are translated
into Prometheus metrics without any labels and with names in which every 
non-alphanumeric character except `_` and `:` is replaced with `_`.

If you have a very large set of metrics you may want to skip the ones that don't
match the mapping configuration. If that is the case you can force this behaviour
using the `-graphite.mapping-strict-match` flag, and it will only store those metrics
you really want.

An example mapping configuration:

```
mappings:
- match: test.dispatcher.*.*.*
  name: dispatcher_events_total
  labels:
    action: $2
    job: test_dispatcher
    outcome: $3
    processor: $1
- match: '*.signup.*.*'
  name: signup_events_total
  labels:
    job: ${1}_server
    outcome: $3
    provider: $2
```

This would transform these example graphite metrics into Prometheus metrics as
follows:

    test.dispatcher.FooProcessor.send.success
     => dispatcher_events_total{processor="FooProcessor", action="send", outcome="success", job="test_dispatcher"}

    foo_product.signup.facebook.failure
     => signup_events_total{provider="facebook", outcome="failure", job="foo_product_server"}

    test.web-server.foo.bar
     => test_web__server_foo_bar{}

### Conversion from legacy configuration

If you have an existing config file using the legacy mapping syntax, you may use [statsd-exporter-convert](https://github.com/bakins/statsd-exporter-convert) to update to the new YAML based syntax.  Here we convert the old example synatx:

```
$ go get -u github.com/bakins/statsd-exporter-convert

$ cat example.conf
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

$ statsd-exporter-convert example.conf
mappings:
- match: test.dispatcher.*.*.*
  name: dispatcher_events_total
  labels:
    action: $2
    job: test_dispatcher
    outcome: $3
    processor: $1
- match: '*.signup.*.*'
  name: signup_events_total
  labels:
    job: ${1}_server
    outcome: $3
    provider: $2
````

## Using Docker

You can deploy this exporter using the [prom/graphite-exporter][hub] Docker image.

For example:

```bash
docker pull prom/graphite-exporter

docker run -d -p 9108:9108 -p 9109:9109 -p 9109:9109/udp \
        -v ${PWD}/graphite_mapping.conf:/tmp/graphite_mapping.conf \
        prom/graphite-exporter --graphite.mapping-config=/tmp/graphite_mapping.conf
```


[circleci]: https://circleci.com/gh/prometheus/graphite_exporter
[hub]: https://hub.docker.com/r/prom/graphite-exporter/
[travis]: https://travis-ci.org/prometheus/graphite_exporter
[quay]: https://quay.io/repository/prometheus/graphite-exporter
