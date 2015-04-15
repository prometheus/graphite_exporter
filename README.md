# Graphite Exporter

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
they are last pushed to. This is configurable with the `--graphite.sample-expiry` flag.
