# Graphite Exporter

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

```sh
make
./graphite_exporter
```

Configure existing monitoring to send Graphite plaintext data to port 9109 on UDP or TCP.
As a simple demonstration:

```sh
echo "test_tcp 1234 $(date +%s)" | nc localhost 9109
echo "test_udp 1234 $(date +%s)" | nc -u -w1 localhost 9109
```

Metrics will be available on [http://localhost:9108/metrics](http://localhost:9108/metrics).

To avoid using unbounded memory, metrics will be garbage collected five minutes after
they are last pushed to. This is configurable with the `--graphite.sample-expiry` flag.

## Graphite Tags

The graphite_exporter accepts metrics in the [tagged carbon format](https://graphite.readthedocs.io/en/latest/tags.html). Labels specified in the mapping configuration take precedence over tags in the metric. In the case where there are valid and invalid tags supplied in one metric, the invalid tags will be dropped and the `graphite_tag_parse_failures` counter will be incremented. The exporter accepts inconsistent label sets, but this may cause issues querying the data in Prometheus.

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

```yaml
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
- match: 'servers\.(.*)\.networking\.subnetworks\.transmissions\.([a-z0-9-]+)\.(.*)'
  match_type: regex
  name: 'servers_networking_transmissions_${3}'
  labels: 
    hostname: ${1}
    device: ${2}
```

This would transform these example graphite metrics into Prometheus metrics as
follows:

```console
test.dispatcher.FooProcessor.send.success
  => dispatcher_events_total{processor="FooProcessor", action="send", outcome="success", job="test_dispatcher"}

foo_product.signup.facebook.failure
  => signup_events_total{provider="facebook", outcome="failure", job="foo_product_server"}

test.web-server.foo.bar
  => test_web__server_foo_bar{}

servers.rack-003-server-c4de.networking.subnetworks.transmissions.eth0.failure.mean_rate
  => servers_networking_transmissions_failure_mean_rate{device="eth0",hostname="rack-003-server-c4de"}
```

### Conversion from legacy configuration

If you have an existing config file using the legacy mapping syntax, you may use [statsd-exporter-convert](https://github.com/bakins/statsd-exporter-convert) to update to the new YAML based syntax.  Here we convert the old example synatx:

```console
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
```

## Using Docker

You can deploy this exporter using the [prom/graphite-exporter][hub] Docker image.

For example:

```bash
docker pull prom/graphite-exporter

docker run -d -p 9108:9108 -p 9109:9109 -p 9109:9109/udp \
        -v ${PWD}/graphite_mapping.conf:/tmp/graphite_mapping.conf \
        prom/graphite-exporter --graphite.mapping-config=/tmp/graphite_mapping.conf
```

## **Experimental**: Importing Whisper data

Import data from Graphite using the bundled `getool`.
See `getool create-blocks --help` for usage.

To import long-term data in a reasonable amount of resources, increase the duration per generated TSDB block.
The `--block-duration` must be a power of two in hours, e.g. `4h`, `8h`, and so on.

To merge the data into an existing Prometheus storage directory, start Prometheus with the `--storage.tsdb.allow-overlapping-blocks` flag.

## Incompatibility with Graphite bridge

This exporter does not work in combination with the [Java client](https://prometheus.github.io/client_java/io/prometheus/client/bridge/Graphite.html) or [Python client](https://github.com/prometheus/client_python#graphite) Graphite bridge.
In the transition to the Graphite data model and back, information is lost.
Additionally, default metrics conflict between the client libraries and the exporter.

Instead, configure Prometheus to scrape your application directly, without the exporter in the middle.
For batch or ephemeral jobs, use the [pushgateway](https://prometheus.io/docs/practices/pushing/) [integration](https://github.com/prometheus/client_python#exporting-to-a-pushgateway).
If you absolutely must push, consider [PushProx](https://github.com/prometheus-community/PushProx) or the [Grafana agent](https://github.com/grafana/agent) instead.

## TLS and basic authentication

Graphite Exporter supports TLS and basic authentication. This enables better control of the various HTTP endpoints.

To use TLS and/or basic authentication, you need to pass a configuration file using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).



[circleci]: https://circleci.com/gh/prometheus/graphite_exporter
[hub]: https://hub.docker.com/r/prom/graphite-exporter/
[travis]: https://travis-ci.org/prometheus/graphite_exporter
[quay]: https://quay.io/repository/prometheus/graphite-exporter
