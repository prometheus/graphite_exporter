## 0.12.4 / 2022-12-05

* [SECURITY] Fix [GHSA-7rg2-cxvp-9p7p](https://github.com/advisories/GHSA-7rg2-cxvp-9p7p) (manual backport of [#209](https://github.com/prometheus/graphite_exporter/pull/209))

## 0.12.3 / 2022-08-06

* [BUGFIX] Fix crash on startup for some configurations ([#198](https://github.com/prometheus/graphite_exporter/pull/198))

For mappings that require backtracking, 0.12.2 would crash on startup due to an uninitialized logger.
If this affected you, consider [changing the order of rules](https://github.com/prometheus/statsd_exporter#ordering-glob-rules) or enabling unordered rules for better performance.

## 0.12.2 / 2022-07-08

* [CHANGE] Update all dependencies ([#193](https://github.com/prometheus/graphite_exporter/pull/193), [#194](https://github.com/prometheus/graphite_exporter/pull/194), [#195](https://github.com/prometheus/graphite_exporter/pull/195), [#196](https://github.com/prometheus/graphite_exporter/pull/196))

This is a comprehensive housekeeping release, bringing all dependencies and the compiler version up to date.

It imports a bug fix in the mapper, allowing metrics with multiple dashes in a row.

## 0.12.1 / 2022-05-06

This is a maintenance release, built with Go 1.17.9 to address security issues.

## 0.12.0 / 2021-12-01

* [FEATURE] Support TLS on web UI and metrics ([#175](https://github.com/prometheus/graphite_exporter/pull/175))

## 0.11.1 / 2021-11-26

* [ENHANCEMENT] Build for windows/arm64 ([#174](https://github.com/prometheus/graphite_exporter/pull/174))

## 0.11.0 / 2021-09-01

* [ENHANCEMENT] Add experimental tool for converting historical data ([#145](https://github.com/prometheus/graphite_exporter/pull/145))

This release adds the `getool` binary to the release tarball.

## 0.10.1 / 2021-05-12

No changes.
This release will include an updated Busybox in the Docker image, which fixes [CVE-2018-1000500](https://nvd.nist.gov/vuln/detail/CVE-2018-1000500).
This security issue does not affect you unless you extend the container and use gzip, but it trips security scanners, so we provide this version.

## 0.10.0 / 2021-04-13

* [CHANGE] Reorganize repository ([#144](https://github.com/prometheus/graphite_exporter/pull/144))
* [ENHANCEMENT] Configuration check ([#146](https://github.com/prometheus/graphite_exporter/pull/146))

The main binary package is now `github.com/prometheus/graphite_exporter/cmd/graphite_exporter`.
This has no effect on those using the binary release.

## 0.9.0 / 2020-07-21

* [ENHANCEMENT] Generate labels from Graphite tags ([#133](https://github.com/prometheus/graphite_exporter/pull/133))

## 0.8.0 / 2020-06-12

* [CHANGE] Update metric mapper and other dependencies ([#127](https://github.com/prometheus/graphite_exporter/pull/127))

This brings the metric mapper to parity with [statsd_exporter 0.16.0](https://github.com/prometheus/statsd_exporter/blob/master/CHANGELOG.md#0160--2020-05-29).
See the statsd exporter changelog for the detailed changes.
Notably, we now support a random-replacement mapping cache.
The changes for the timer type configuration do not affect this exporter as Graphite only supports gauge-type metrics.

## 0.7.1 / 2020-05-12

* [BUGFIX] Fix "superfluous response.WriteHeader call" through dependency update ([#125](https://github.com/prometheus/graphite_exporter/pull/125))

## 0.7.0 / 2020-02-28

* [CHANGE] Update logging library and flags ([#109](https://github.com/prometheus/graphite_exporter/pull/109))
* [CHANGE] Updated prometheus golang client and statsd mapper dependency. ([#113](https://github.com/prometheus/graphite_exporter/pull/113))

This release updates several dependencies. Logging-related flags have changed.

The metric mapping library is now at the level of [statsd exporter 0.14.1](https://github.com/prometheus/statsd_exporter/blob/master/CHANGELOG.md#0141--2010-01-13), bringing in various performance improvements. See the statsd exporter changelog for the detailed changes.

## 0.6.2 / 2019-06-03

* [CHANGE] Do not run as root in the Docker container by default ([#85](https://github.com/prometheus/graphite_exporter/pull/85))
* [BUGFIX] Serialize processing of samples ([#94](https://github.com/prometheus/graphite_exporter/pull/94))

This issue fixes a race condition in sample processing that showed if multiple
clients sent metrics simultaneously, or multiple metrics were sent in
individual UDP packets. It would manifest as duplicate metrics being exported
(0.4.x) or the metrics endpoint failing altogether (0.5.0).

## 0.5.0 / 2019-02-28

* [ENHANCEMENT] Accept 'name' as a label ([#75](https://github.com/prometheus/graphite_exporter/pull/75))
* [BUGFIX] Update the mapper to fix captures being clobbered ([#77](https://github.com/prometheus/graphite_exporter/pull/77))
* [BUGFIX] Do not mask the pprof endpoints ([#67](https://github.com/prometheus/graphite_exporter/pull/67))

This release also pulls in a more recent version of the Prometheus client library with improved validation and performance.

## 0.4.2 / 2018-11-26

* [BUGFIX] Fix segfault in mapper if mapping config is provided ([#63](https://github.com/prometheus/graphite_exporter/pull/63))

## 0.4.1 / 2018-11-23

No changes.

## 0.4.0 / 2018-11-23

* [ENHANCEMENT] Log incoming and parsed samples if debug logging is enabled ([#58](https://github.com/prometheus/graphite_exporter/pull/58))
* [ENHANCEMENT] Speed up glob matching ([#59](https://github.com/prometheus/graphite_exporter/pull/59))

This release replaces the implementation of the glob matching mechanism,
speeding it up significantly. In certain sub-optimal configurations, a warning
is logged.

This major enhancement was contributed by Wangchong Zhou in [prometheus/statsd_exporter#157](https://github.com/prometheus/statsd_exporter/pulls/157).

## 0.3.0 / 2018-08-22

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

## 0.2.0 / 2017-03-01

* [FEATURE] Added flag to allow dropping of unmatched metrics
* [ENHANCEMENT] Logging changes and standardisation

## 0.1.0 / 2015-05-05

Initial release.
