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

package line

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/statsd_exporter/pkg/mapper"

	"github.com/prometheus/graphite_exporter/pkg/graphitesample"
	"github.com/prometheus/graphite_exporter/pkg/metricmapper"
)

var (
	invalidMetricChars  = regexp.MustCompile("[^a-zA-Z0-9_:]")
	metricNameKeysIndex = newMetricNameAndKeys()
)

// metricNameAndKeys is a cache of metric names and the label keys previously used
type metricNameAndKeys struct {
	mtx   sync.Mutex
	cache map[string]string
}

func newMetricNameAndKeys() *metricNameAndKeys {
	x := metricNameAndKeys{
		cache: make(map[string]string),
	}

	return &x
}

func keysFromLabels(labels prometheus.Labels) string {
	labelKeys := make([]string, len(labels))
	for k := range labels {
		labelKeys = append(labelKeys, k)
	}

	sort.Strings(labelKeys)

	return strings.Join(labelKeys, ",")
}

// checkNameAndKeys returns true if metric has the same label keys or is new, false if not
func (c *metricNameAndKeys) checkNameAndKeys(name string, labels prometheus.Labels) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	providedKeys := keysFromLabels(labels)

	if keys, found := c.cache[name]; found {
		return keys == providedKeys
	}

	c.cache[name] = providedKeys

	return true
}

func parseMetricNameAndTags(name string, labels prometheus.Labels, tagErrors prometheus.Counter) (string, error) {
	if strings.ContainsRune(name, ';') {
		// name contains tags - parse tags and add to labels
		if strings.Count(name, ";") != strings.Count(name, "=") {
			tagErrors.Inc()
			return name, fmt.Errorf("error parsing tags on %s", name)
		}

		parts := strings.Split(name, ";")
		parsedName := parts[0]
		tags := parts[1:]

		for _, tag := range tags {
			kv := strings.SplitN(tag, "=", 2)
			if len(kv) != 2 {
				// we may have added bad labels already...
				tagErrors.Inc()
				return name, fmt.Errorf("error parsing tags on %s", name)
			}

			k := kv[0]
			v := kv[1]
			labels[k] = v
		}

		return parsedName, nil
	}

	return name, nil
}

// ProcessLine takes a graphite metric line as a string, processes it into a GraphiteSample, and sends it to the sample channel
func ProcessLine(line string, metricmapper metricmapper.MetricMapper, sampleCh chan<- *graphitesample.GraphiteSample, strictMatch bool, tagErrors prometheus.Counter, lastProcessed prometheus.Gauge, invalidMetrics prometheus.Counter, logger log.Logger) {
	line = strings.TrimSpace(line)
	level.Debug(logger).Log("msg", "Incoming line", "line", line)
	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		level.Info(logger).Log("msg", "Invalid part count", "parts", len(parts), "line", line)
		return
	}

	originalName := parts[0]

	labels := make(prometheus.Labels)
	name, err := parseMetricNameAndTags(originalName, labels, tagErrors)
	if err != nil {
		level.Info(logger).Log("msg", "Invalid tags", "line", line)
		return
	}

	// check to ensure the same tags are present
	if validKeys := metricNameKeysIndex.checkNameAndKeys(name, labels); !validKeys {
		level.Info(logger).Log("msg", "Dropped because metric keys do not match previously used keys", "line", line)
		invalidMetrics.Inc()

		return
	}

	mapping, mlabels, mappingPresent := metricmapper.GetMapping(name, mapper.MetricTypeGauge)

	if (mappingPresent && mapping.Action == mapper.ActionTypeDrop) || (!mappingPresent && strictMatch) {
		return
	}

	if mappingPresent {
		name = invalidMetricChars.ReplaceAllString(mapping.Name, "_")

		// append labels from the mapping to those parsed, with mapping labels overriding
		for k, v := range mlabels {
			labels[k] = v
		}
	} else {
		name = invalidMetricChars.ReplaceAllString(name, "_")
	}

	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		level.Info(logger).Log("msg", "Invalid value", "line", line)
		return
	}

	timestamp, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		level.Info(logger).Log("msg", "Invalid timestamp", "line", line)
		return
	}

	sample := graphitesample.GraphiteSample{
		OriginalName: originalName,
		Name:         name,
		Value:        value,
		Labels:       labels,
		Type:         prometheus.GaugeValue,
		Help:         fmt.Sprintf("Graphite metric %s", name),
		Timestamp:    time.Unix(int64(timestamp), int64(math.Mod(timestamp, 1.0)*1e9)),
	}
	level.Debug(logger).Log("msg", "Processing sample", "sample", sample)
	lastProcessed.Set(float64(time.Now().UnixNano()) / 1e9)
	sampleCh <- &sample
}
