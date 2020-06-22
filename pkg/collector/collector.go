// Copyright 2020 The Prometheus Authors
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
	"io"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/graphite_exporter/pkg/graphitesample"
	"github.com/prometheus/graphite_exporter/pkg/line"
	"github.com/prometheus/graphite_exporter/pkg/metricmapper"
)

type graphiteCollector struct {
	Mapper             metricmapper.MetricMapper
	Samples            map[string]*graphitesample.GraphiteSample
	mu                 *sync.Mutex
	SampleCh           chan *graphitesample.GraphiteSample
	lineCh             chan string
	StrictMatch        bool
	sampleExpiry       time.Duration
	Logger             log.Logger
	tagErrors          prometheus.Counter
	lastProcessed      prometheus.Gauge
	sampleExpiryMetric prometheus.Gauge
	invalidMetrics     prometheus.Counter
}

func NewGraphiteCollector(logger log.Logger, strictMatch bool, sampleExpiry time.Duration, tagErrors prometheus.Counter, lastProcessed prometheus.Gauge, sampleExpiryMetric prometheus.Gauge, invalidMetrics prometheus.Counter) *graphiteCollector {
	c := &graphiteCollector{
		SampleCh:           make(chan *graphitesample.GraphiteSample),
		lineCh:             make(chan string),
		mu:                 &sync.Mutex{},
		Samples:            map[string]*graphitesample.GraphiteSample{},
		StrictMatch:        strictMatch,
		sampleExpiry:       sampleExpiry,
		Logger:             logger,
		tagErrors:          tagErrors,
		lastProcessed:      lastProcessed,
		sampleExpiryMetric: sampleExpiryMetric,
		invalidMetrics:     invalidMetrics,
	}

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

func (c *graphiteCollector) processLines() {
	for l := range c.lineCh {
		line.ProcessLine(l, c.Mapper, c.SampleCh, c.StrictMatch, c.tagErrors, c.lastProcessed, c.invalidMetrics, c.Logger)
	}
}

func (c *graphiteCollector) processSamples() {
	ticker := time.NewTicker(time.Minute).C

	for {
		select {
		case sample, ok := <-c.SampleCh:
			if sample == nil || !ok {
				return
			}

			c.mu.Lock()
			c.Samples[sample.OriginalName] = sample
			c.mu.Unlock()
		case <-ticker:
			// Garbage collect expired Samples.
			ageLimit := time.Now().Add(-c.sampleExpiry)

			c.mu.Lock()
			for k, sample := range c.Samples {
				if ageLimit.After(sample.Timestamp) {
					delete(c.Samples, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Collect implements prometheus.Collector.
func (c graphiteCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- c.lastProcessed

	c.mu.Lock()
	level.Debug(c.Logger).Log("msg", "Samples length", "len", len(c.Samples))
	samples := make([]*graphitesample.GraphiteSample, 0, len(c.Samples))

	for _, sample := range c.Samples {
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

// Describe implements prometheus.Collector.
func (c graphiteCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.lastProcessed.Desc()
}
