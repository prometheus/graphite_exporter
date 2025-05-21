// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"math"
	_ "net/http/pprof"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
)

var invalidMetricChars = regexp.MustCompile("[^a-zA-Z0-9_:]")

type graphiteCollector struct {
	samples            map[string]*graphiteSample
	mu                 *sync.Mutex
	mapper             metricMapper
	sampleCh           chan *graphiteSample
	lineCh             chan string
	strictMatch        bool
	logger             *slog.Logger
	tagParseFailures   prometheus.Counter
	lastProcessed      prometheus.Gauge
	sampleExpiryMetric prometheus.Gauge
	sampleExpiry       time.Duration
}

func NewGraphiteCollector(logger *slog.Logger, strictMatch bool, sampleExpiry time.Duration) *graphiteCollector {
	c := &graphiteCollector{
		sampleCh:    make(chan *graphiteSample),
		lineCh:      make(chan string),
		mu:          &sync.Mutex{},
		samples:     map[string]*graphiteSample{},
		strictMatch: strictMatch,
		logger:      logger,
		tagParseFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "graphite_tag_parse_failures",
				Help: "Total count of samples with invalid tags",
			}),
		lastProcessed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "graphite_last_processed_timestamp_seconds",
				Help: "Unix timestamp of the last processed graphite metric.",
			},
		),
		sampleExpiry: sampleExpiry,
		sampleExpiryMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "graphite_sample_expiry_seconds",
				Help: "How long in seconds a metric sample is valid for.",
			},
		),
	}
	c.sampleExpiryMetric.Set(sampleExpiry.Seconds())
	go c.processSamples()
	go c.processLines()
	return c
}

func (c *graphiteCollector) ProcessReader(reader io.Reader) {
	lineScanner := bufio.NewScanner(reader)
	for {
		if ok := lineScanner.Scan(); !ok {
			break
		}
		c.lineCh <- lineScanner.Text()
	}
}

func (c *graphiteCollector) SetMapper(m metricMapper) {
	c.mapper = m
}

func (c *graphiteCollector) processLines() {
	for line := range c.lineCh {
		c.processLine(line)
	}
}

func (c *graphiteCollector) parseMetricNameAndTags(name string) (string, prometheus.Labels, error) {
	var err error

	labels := make(prometheus.Labels)

	parts := strings.Split(name, ";")
	parsedName := parts[0]

	tags := parts[1:]
	for _, tag := range tags {
		kv := strings.SplitN(tag, "=", 2)
		if len(kv) != 2 {
			// don't add this tag, continue processing tags but return an error
			c.tagParseFailures.Inc()
			err = fmt.Errorf("error parsing tag %s", tag)
			continue
		}

		k := kv[0]
		v := kv[1]
		labels[k] = v
	}

	return parsedName, labels, err
}

func (c *graphiteCollector) processLine(line string) {
	line = strings.TrimSpace(line)
	c.logger.Debug("Incoming line", "line", line)

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		c.logger.Info("Invalid part count", "parts", len(parts), "line", line)
		return
	}

	originalName := parts[0]

	parsedName, labels, err := c.parseMetricNameAndTags(originalName)
	if err != nil {
		c.logger.Debug("Invalid tags", "line", line, "err", err.Error())
	}

	mapping, mappingLabels, mappingPresent := c.mapper.GetMapping(parsedName, mapper.MetricTypeGauge)

	// add mapping labels to parsed labels
	for k, v := range mappingLabels {
		labels[k] = v
	}

	if (mappingPresent && mapping.Action == mapper.ActionTypeDrop) || (!mappingPresent && c.strictMatch) {
		return
	}

	var name string
	if mappingPresent {
		name = invalidMetricChars.ReplaceAllString(mapping.Name, "_")
	} else {
		name = invalidMetricChars.ReplaceAllString(parsedName, "_")
	}

	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		c.logger.Info("Invalid value", "line", line)
		return
	}
	if mappingPresent && mapping.Scale.Set {
		value *= mapping.Scale.Val
	}

	timestamp, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		c.logger.Info("Invalid timestamp", "line", line)
		return
	}
	sample := graphiteSample{
		OriginalName: originalName,
		Name:         name,
		Value:        value,
		Labels:       labels,
		Type:         prometheus.GaugeValue,
		Help:         fmt.Sprintf("Graphite metric %s", name),
		Timestamp:    time.Unix(int64(timestamp), int64(math.Mod(timestamp, 1.0)*1e9)),
	}
	c.logger.Debug("Processing sample", "sample", sample)
	c.lastProcessed.Set(float64(time.Now().UnixNano()) / 1e9)
	c.sampleCh <- &sample
}

func (c *graphiteCollector) processSamples() {
	ticker := time.NewTicker(time.Minute).C

	for {
		select {
		case sample, ok := <-c.sampleCh:
			if sample == nil || !ok {
				return
			}
			c.mu.Lock()
			c.samples[sample.OriginalName] = sample
			c.mu.Unlock()
		case <-ticker:
			// Garbage collect expired samples.
			ageLimit := time.Now().Add(-c.sampleExpiry)
			c.mu.Lock()
			for k, sample := range c.samples {
				if ageLimit.After(sample.Timestamp) {
					delete(c.samples, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Collect implements prometheus.Collector.
func (c graphiteCollector) Collect(ch chan<- prometheus.Metric) {
	c.lastProcessed.Collect(ch)
	c.sampleExpiryMetric.Collect(ch)
	c.tagParseFailures.Collect(ch)

	c.mu.Lock()
	samples := make([]*graphiteSample, 0, len(c.samples))
	for _, sample := range c.samples {
		samples = append(samples, sample)
	}
	c.mu.Unlock()

	ageLimit := time.Now().Add(-c.sampleExpiry)
	for _, sample := range samples {
		if ageLimit.After(sample.Timestamp) {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(sample.Name, sample.Help, []string{}, sample.Labels),
			sample.Type,
			sample.Value,
		)
	}
}

// Describe implements prometheus.Collector but does not yield a description
// for Graphite metrics, allowing inconsistent label sets
func (c graphiteCollector) Describe(ch chan<- *prometheus.Desc) {
	c.lastProcessed.Describe(ch)
	c.sampleExpiryMetric.Describe(ch)
	c.tagParseFailures.Describe(ch)
}

type graphiteSample struct {
	OriginalName string
	Name         string
	Labels       prometheus.Labels
	Help         string
	Value        float64
	Type         prometheus.ValueType
	Timestamp    time.Time
}

func (s graphiteSample) String() string {
	return fmt.Sprintf("%#v", s)
}

type metricMapper interface {
	GetMapping(string, mapper.MetricType) (*mapper.MetricMapping, prometheus.Labels, bool)
	InitFromFile(string) error
}
