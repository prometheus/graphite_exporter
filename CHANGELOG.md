# 0.6.2 / 2019-06-03

* [CHANGE] Do not run as root in the Docker container by default ([#85](https://github.com/prometheus/graphite_exporter/pull/85))
* [BUGFIX] Serialize processing of samples ([#94](https://github.com/prometheus/graphite_exporter/pull/94))

This issue fixes a race condition in sample processing that showed if multiple
clients sent metrics simultaneously, or multiple metrics were sent in
individual UDP packets. It would manifest as duplicate metrics being exported
(0.4.x) or the metrics endpoint failing altogether (0.5.0).

# 0.5.0 / 2019-02-28

* [BUGFIX] Update the mapper to fix captures being clobbered ([#77](https://github.com/prometheus/graphite_exporter/pull/77))
* [BUGFIX] Do not mask the pprof endpoints ([#67](https://github.com/prometheus/graphite_exporter/pull/67))
* [ENHANCEMENT] Accept 'name' as a label ([#75](https://github.com/prometheus/graphite_exporter/pull/75))

This release also pulls in a more recent version of the Prometheus client library with improved validation and performance.

# 0.4.2 / 2018-11-26

* [BUGFIX] Fix segfault in mapper if mapping config is provided ([#63](https://github.com/prometheus/graphite_exporter/pull/63))

# 0.4.1 / 2018-11-23

No changes.

# 0.4.0 / 2018-11-23

* [ENHANCEMENT] Log incoming and parsed samples if debug logging is enabled ([#58](https://github.com/prometheus/graphite_exporter/pull/58))
* [ENHANCEMENT] Speed up glob matching ([#59](https://github.com/prometheus/graphite_exporter/pull/59))

This release replaces the implementation of the glob matching mechanism,
speeding it up significantly. In certain sub-optimal configurations, a warning
is logged.

This major enhancement was contributed by Wangchong Zhou in [prometheus/statsd_exporter#157](https://github.com/prometheus/statsd_exporter/pulls/157).

# 0.3.0 / 2018-08-22

This release contains two major breaking changes:

Flags now require two dashes (`--help` instead of `-help`).

The configuration format is now YAML, and uses the same format as the [statsd exporter](https://github.com/prometheus/statsd_exporter), minus support for
metric types other than gauges.
There is a [conversion tool](https://github.com/bakins/statsd-exporter-convert) available.
This change adds new features to the mappings:
It is now possible to specify the "name" label.
Regular expressions can be used to match on Graphite metric names beyond extracting dot-separated components.

* [CHANGE] Use YAML configuration format and mapper from statsd exporter ([#52](https://github.com/prometheus/graphite_exporter/pull/52))
* [CHANGE] Switch to the Kingpin flag library ([#30](https://github.com/prometheus/graphite_exporter/30))
* [FEATURE] Add metric for the sample expiry setting ([#34](https://github.com/prometheus/graphite_exporter/34))
* [FEATURE] Add pprof endpoint ([#33](https://github.com/prometheus/graphite_exporter/33))
* [BUGFIX] Accept whitespace around the Graphite protocol lines ([#53](https://github.com/prometheus/graphite_exporter/53))

# 0.2.0 / 2017-03-01

* [FEATURE] Added flag to allow dropping of unmatched metrics
* [ENHANCEMENT] Logging changes and standardisation


# 0.1.0 / 2015-05-05

Initial release.
