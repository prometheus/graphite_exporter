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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
	"github.com/stretchr/testify/assert"
)

func TestParseNameAndTags(t *testing.T) {
	logger := promslog.NewNopLogger()
	c := NewGraphiteCollector(logger, false, 5*time.Minute)
	type testCase struct {
		line       string
		parsedName string
		labels     prometheus.Labels
		willError  bool
	}

	testCases := map[string]testCase{
		"good tags": {
			line:       "my_simple_metric_with_tags;tag1=value1;tag2=value2",
			parsedName: "my_simple_metric_with_tags",
			labels: prometheus.Labels{
				"tag1": "value1",
				"tag2": "value2",
			},
		},
		"no tag value": {
			line:       "my_simple_metric_with_bad_tags;tag1=value3;tag2",
			parsedName: "my_simple_metric_with_bad_tags",
			labels: prometheus.Labels{
				"tag1": "value3",
			},
			willError: true,
		},
		"no tag value in middle": {
			line:       "my_simple_metric_with_bad_tags;tag1=value3;tag2;tag3=value4",
			parsedName: "my_simple_metric_with_bad_tags",
			labels: prometheus.Labels{
				"tag1": "value3",
				"tag3": "value4",
			},
			willError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			n, parsedLabels, err := c.parseMetricNameAndTags(testCase.line)
			if !testCase.willError {
				assert.NoError(t, err, "Got unexpected error parsing %s", testCase.line)
			}
			assert.Equal(t, testCase.parsedName, n)
			assert.Equal(t, testCase.labels, parsedLabels)
		})
	}
}

func TestProcessLine(t *testing.T) {
	type testCase struct {
		line           string
		name           string
		mappingLabels  prometheus.Labels
		sampleLabels   prometheus.Labels
		value          float64
		mappingPresent bool
		willFail       bool
		action         mapper.ActionType
		strict         bool
		scale          mapper.MaybeFloat64
	}

	testCases := map[string]testCase{
		"simple metric": {
			line: "my.simple.metric 9001 1534620625",
			name: "my_simple_metric",
			mappingLabels: prometheus.Labels{
				"foo":  "bar",
				"zip":  "zot",
				"name": "alabel",
			},
			sampleLabels: prometheus.Labels{
				"foo":  "bar",
				"zip":  "zot",
				"name": "alabel",
			},
			mappingPresent: true,
			value:          float64(9001),
		},
		"existing metric with different labels is accepted": {
			line: "my.simple.metric.baz 9002 1534620625",
			name: "my_simple_metric",
			mappingLabels: prometheus.Labels{
				"baz": "bat",
			},
			sampleLabels: prometheus.Labels{
				"baz": "bat",
			},
			mappingPresent: true,
			value:          float64(9002),
			willFail:       false,
		},
		"mapped metric": {
			line: "my.simple.metric.new.baz 9002 1534620625",
			name: "my_simple_metric_new",
			mappingLabels: prometheus.Labels{
				"baz": "bat",
			},
			sampleLabels: prometheus.Labels{
				"baz": "bat",
			},
			mappingPresent: true,
			value:          float64(9002),
		},
		"no mapping metric": {
			line:           "my.nomap.metric 9001 1534620625",
			name:           "my_nomap_metric",
			value:          float64(9001),
			sampleLabels:   prometheus.Labels{},
			mappingPresent: false,
		},
		"no mapping metric with no value": {
			line:     "my.nomap.metric.novalue 9001 ",
			name:     "my_nomap_metric_novalue",
			value:    float64(9001),
			willFail: true,
		},
		"mapping type drop": {
			line:           "my.mapped.metric.drop 55 1534620625",
			name:           "my_mapped_metric_drop",
			mappingPresent: true,
			willFail:       true,
			action:         mapper.ActionTypeDrop,
		},
		"strict mapped metric": {
			line:           "my.mapped.strict.metric 55 1534620625",
			name:           "my_mapped_strict_metric",
			value:          float64(55),
			mappingPresent: true,
			mappingLabels:  prometheus.Labels{},
			sampleLabels:   prometheus.Labels{},
			willFail:       false,
			strict:         true,
		},
		"strict unmapped metric will drop": {
			line:           "my.mapped.strict.metric.drop 55 1534620625",
			name:           "my_mapped_strict_metric_drop",
			mappingPresent: false,
			willFail:       true,
			strict:         true,
		},
		"unmapped metric with tags": {
			line: "my.simple.metric.with.tags;tag1=value1;tag2=value2 9002 1534620625",
			name: "my_simple_metric_with_tags",
			sampleLabels: prometheus.Labels{
				"tag1": "value1",
				"tag2": "value2",
			},
			mappingPresent: false,
			value:          float64(9002),
		},
		"unmapped metric with different tag values": {
			// same tags, different values, should parse
			line: "my.simple.metric.with.tags;tag1=value3;tag2=value4 9002 1534620625",
			name: "my_simple_metric_with_tags",
			sampleLabels: prometheus.Labels{
				"tag1": "value3",
				"tag2": "value4",
			},
			mappingPresent: false,
			value:          float64(9002),
		},
		"mapping labels added to tags": {
			// labels in mapping should be added to sample labels
			line: "my.mapped.metric.with.tags;tag1=value3;tag2=value4 9003 1534620625",
			name: "my_mapped_metric_with_tags",
			mappingLabels: prometheus.Labels{
				"foobar": "baz",
			},
			sampleLabels: prometheus.Labels{
				"tag1":   "value3",
				"tag2":   "value4",
				"foobar": "baz",
			},
			mappingPresent: true,
			value:          float64(9003),
		},
		"scaled metric": {
			line: "my.scaled.metric 9001 1534620625",
			name: "my_scaled_metric",
			scale: mapper.MaybeFloat64{
				Set: true,
				Val: 1024,
			},
			sampleLabels:   prometheus.Labels{},
			mappingPresent: true,
			value:          float64(9001) * 1024,
		},
	}

	c := NewGraphiteCollector(promslog.NewNopLogger(), false, 5*time.Minute)

	for _, testCase := range testCases {
		if testCase.mappingPresent {
			c.mapper = &mockMapper{
				name:    testCase.name,
				labels:  testCase.mappingLabels,
				action:  testCase.action,
				present: testCase.mappingPresent,
				scale:   testCase.scale,
			}
		} else {
			c.mapper = &mockMapper{
				present: testCase.mappingPresent,
			}
		}

		c.strictMatch = testCase.strict
		c.processLine(testCase.line)
	}

	c.sampleCh <- nil
	for name, k := range testCases {
		t.Run(name, func(t *testing.T) {
			originalName, _, _ := strings.Cut(k.line, " ")
			sample := sampleByOriginalName(c.samples, originalName)
			if k.willFail {
				assert.Nil(t, sample, "Found %s", k.name)
			} else {
				if assert.NotNil(t, sample, "Missing %s", k.name) {
					assert.Equal(t, k.name, sample.Name)
					assert.Equal(t, k.sampleLabels, sample.Labels)
					assert.Equal(t, k.value, sample.Value)
				}
			}
		})
	}
}

func TestCollectCollidingGraphiteNames(t *testing.T) {
	now := time.Now().Unix()
	for name, tc := range map[string]struct {
		firstLine  string
		secondLine string
		wantValue  float64
	}{
		"later timestamp wins": {
			firstLine:  fmt.Sprintf("graphite.name.one 1 %d", now-1),
			secondLine: fmt.Sprintf("graphite.name.two 2 %d", now),
			wantValue:  2,
		},
		"later timestamp wins when it arrives first": {
			firstLine:  fmt.Sprintf("graphite.name.one 2 %d", now),
			secondLine: fmt.Sprintf("graphite.name.two 1 %d", now-1),
			wantValue:  2,
		},
		"equal timestamps keep the incoming sample": {
			firstLine:  fmt.Sprintf("graphite.name.one 1 %d", now),
			secondLine: fmt.Sprintf("graphite.name.two 2 %d", now),
			wantValue:  2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := NewGraphiteCollector(promslog.NewNopLogger(), false, 5*time.Minute)
			c.mapper = &mockMapper{
				name: "colliding_metric",
				labels: prometheus.Labels{
					"foo": "bar",
				},
				present: true,
			}

			c.processLine(tc.firstLine)
			c.processLine(tc.secondLine)
			c.sampleCh <- nil

			assert.Len(t, c.samples, 1)

			family := gatherMetricFamily(t, c, "colliding_metric")
			if assert.NotNil(t, family) {
				assert.Len(t, family.Metric, 1)
				assert.Equal(t, tc.wantValue, family.Metric[0].GetGauge().GetValue())
			}
		})
	}
}

func TestCollectDistinctIdentities(t *testing.T) {
	now := time.Now().Unix()
	c := NewGraphiteCollector(promslog.NewNopLogger(), false, 5*time.Minute)
	c.mapper = &mockMapper{present: false}

	c.processLine(fmt.Sprintf("metric.one;env=prod 1 %d", now))
	c.processLine(fmt.Sprintf("metric.two;env=dev 2 %d", now))
	c.sampleCh <- nil

	assert.Len(t, c.samples, 2)

	families := gatherMetricFamilies(t, c)
	one := families["metric_one"]
	two := families["metric_two"]
	if assert.NotNil(t, one) {
		assert.Len(t, one.Metric, 1)
		assert.Equal(t, 1.0, one.Metric[0].GetGauge().GetValue())
	}
	if assert.NotNil(t, two) {
		assert.Len(t, two.Metric, 1)
		assert.Equal(t, 2.0, two.Metric[0].GetGauge().GetValue())
	}
}

func sampleByOriginalName(samples map[string]*graphiteSample, originalName string) *graphiteSample {
	for _, sample := range samples {
		if sample.OriginalName == originalName {
			return sample
		}
	}
	return nil
}

func gatherMetricFamily(t *testing.T, c *GraphiteCollector, name string) *dto.MetricFamily {
	t.Helper()
	return gatherMetricFamilies(t, c)[name]
}

func gatherMetricFamilies(t *testing.T, c *GraphiteCollector) map[string]*dto.MetricFamily {
	t.Helper()
	reg := prometheus.NewRegistry()
	assert.NoError(t, reg.Register(c))
	families, err := reg.Gather()
	assert.NoError(t, err)
	out := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		out[family.GetName()] = family
	}
	return out
}
