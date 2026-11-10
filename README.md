# Graphite Exporter

[![Build Status](https://github.com/prometheus/graphite_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/prometheus/graphite_exporter/actions/workflows/ci.yml)
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

The graphite_exporter accepts metrics in the [tagged carbon format](https://graphite.readthedocs.io/en/latest/tags.html). In the case where there are valid and invalid tags supplied in one metric, the invalid tags will be dropped and the `graphite_tag_parse_failures` counter will be incremented. The exporter accepts inconsistent label sets, but this may cause issues querying the data in Prometheus.

By default, labels explicitly specified in configuration take precedence over labels from the metric. To set the label from the metric instead, use [`honor_labels`](https://github.com/prometheus/statsd_exporter/#honor-labels).


## Metric Mapping and Configuration

**Please note there has been a breaking change in configuration after version 0.2.0.  The YAML style config from [statsd_exporter](https://github.com/prometheus/statsd_exporter) is now used.  See conversion instructions below**

### YAML Config

The graphite_exporter can be configured to translate specific dot-separated
Graphite metrics into labeled Prometheus metrics via a YAML configuration file.
This file shares syntax and logic with
[statsd_exporter](https://github.com/prometheus/statsd_exporter). However,
graphite_exporter does not support all parsing features. Features based on the
`timer_type` option do not function. Regex matching, capture groups, and map/drop
actions work as described below.

#### Glob matching

The default and fastest match type is `glob`. Each `*` captures one dot-separated
component, which can be referenced as `$1`, `$2`, and so on in the metric name or
labels.

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
```

These mappings transform Graphite metrics into Prometheus metrics as
follows:

```console
test.dispatcher.FooProcessor.send.success
  => dispatcher_events_total{processor="FooProcessor", action="send", outcome="success", job="test_dispatcher"}

foo_product.signup.facebook.failure
  => signup_events_total{provider="facebook", outcome="failure", job="foo_product_server"}
```

#### Regular expression matching

Use `match_type: regex` when glob matching cannot express the required structure.
Capture groups use the same `$1` or `${1}` references as glob mappings. Single
quotes around the YAML value keep regular expression backslashes literal.

Glob mappings are always evaluated before regex mappings, regardless of their
order in the file. Regex mappings are then evaluated in order, and the first
match wins. Prefer glob mappings where possible because regex mappings are
evaluated one by one.

For example:

```yaml
mappings:
- match: '^servers\.([^.]+)\.networking\.subnetworks\.transmissions\.([a-z0-9-]+)\.(.+)$'
  match_type: regex
  name: 'servers_networking_transmissions_${3}'
  labels:
    hostname: '${1}'
    device: '${2}'
```

This mapping produces:

```console
servers.rack-003-server-c4de.networking.subnetworks.transmissions.eth0.failure.mean_rate
  => servers_networking_transmissions_failure_mean_rate{device="eth0", hostname="rack-003-server-c4de"}
```

#### Handling unmatched metrics

By default, metrics that do not match any mapping are still exported without
labels. Every character in their name other than letters, digits, `_`, and `:`
is replaced with `_`:

```console
test.web-server.foo.bar
  => test_web_server_foo_bar{}
```

To drop every unmatched metric, start graphite_exporter with
`--graphite.mapping-strict-match`. Metrics dropped by strict matching increment
`graphite_dropped_samples_total`.

During a mapping rollout, a final regex catch-all can retain unmatched metrics
and expose their original Graphite names:

```yaml
mappings:
# Add specific glob and regex mappings above this rule.
- match: '.+'
  match_type: regex
  name: '$0'
  labels:
    graphite_metric_name: '$0'
```

For example, `test.web-server.foo.bar` is exported as
`test_web_server_foo_bar{graphite_metric_name="test.web-server.foo.bar"}`. The
catch-all must be the last regex mapping. It also counts as a match when
`--graphite.mapping-strict-match` is enabled, so these metrics are retained rather
than dropped. The extra label creates one label value per original metric name;
consider using this rule temporarily when the input has high cardinality.

To explicitly drop everything not selected by earlier mappings, use a final
drop rule instead:

```yaml
mappings:
# Add specific glob and regex mappings above this rule.
- match: '.+'
  match_type: regex
  action: drop
  name: dropped
```

Explicitly dropped metrics also increment `graphite_dropped_samples_total`.

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
