// Copyright 2018 The Prometheus Authors
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

package main

import (
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
	"github.com/stretchr/testify/assert"
)

type mockMapper struct {
	labels  prometheus.Labels
	present bool
	name    string
	action  mapper.ActionType
}

func (m *mockMapper) GetMapping(metricName string, metricType mapper.MetricType) (*mapper.MetricMapping, prometheus.Labels, bool) {

	mapping := mapper.MetricMapping{Name: m.name, Action: m.action}

	return &mapping, m.labels, m.present

}

func (m *mockMapper) InitFromFile(string, int, ...mapper.CacheOption) error {
	return nil
}
func (m *mockMapper) InitCache(int, ...mapper.CacheOption) {

}

func TestParseNameAndTags(t *testing.T) {
	type testCase struct {
		line       string
		parsedName string
		labels     prometheus.Labels
		willError  bool
	}

	testCases := []testCase{
		{
			line:       "my_simple_metric_with_tags;tag1=value1;tag2=value2",
			parsedName: "my_simple_metric_with_tags",
			labels: prometheus.Labels{
				"tag1": "value1",
				"tag2": "value2",
			},
		},
		{
			line:       "my_simple_metric_with_bad_tags;tag1=value3;tag2",
			parsedName: "my_simple_metric_with_bad_tags",
			labels: prometheus.Labels{
				"tag1": "value3",
			},
			willError: true,
		},
		{
			line:       "my_simple_metric_with_bad_tags;tag1=value3;tag2;tag3=value4",
			parsedName: "my_simple_metric_with_bad_tags",
			labels: prometheus.Labels{
				"tag1": "value3",
				"tag3": "value4",
			},
			willError: true,
		},
	}

	for _, testCase := range testCases {
		n, parsedLabels, err := parseMetricNameAndTags(testCase.line)
		if !testCase.willError {
			assert.NoError(t, err, "Got unexpected error parsing %s", testCase.line)
		}
		assert.Equal(t, testCase.parsedName, n)
		assert.Equal(t, testCase.labels, parsedLabels)
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
	}

	testCases := []testCase{
		{
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
		{
			// will fail since my_simple_metric has different label keys than in the previous test case
			line: "my.simple.metric.baz 9002 1534620625",
			name: "my_simple_metric",
			mappingLabels: prometheus.Labels{
				"baz": "bat",
			},
			mappingPresent: true,
			value:          float64(9002),
			willFail:       true,
		},
		{
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
		{
			line:           "my.nomap.metric 9001 1534620625",
			name:           "my_nomap_metric",
			value:          float64(9001),
			sampleLabels:   prometheus.Labels{},
			mappingPresent: false,
		},
		{
			line:     "my.nomap.metric.novalue 9001 ",
			name:     "my_nomap_metric_novalue",
			value:    float64(9001),
			willFail: true,
		},
		{
			line:           "my.mapped.metric.drop 55 1534620625",
			name:           "my_mapped_metric_drop",
			mappingPresent: true,
			willFail:       true,
			action:         mapper.ActionTypeDrop,
		},
		{
			line:           "my.mapped.strict.metric 55 1534620625",
			name:           "my_mapped_strict_metric",
			value:          float64(55),
			mappingPresent: true,
			mappingLabels:  prometheus.Labels{},
			sampleLabels:   prometheus.Labels{},
			willFail:       false,
			strict:         true,
		},
		{
			line:           "my.mapped.strict.metric.drop 55 1534620625",
			name:           "my_mapped_strict_metric_drop",
			mappingPresent: false,
			willFail:       true,
			strict:         true,
		},
		{
			line: "my.simple.metric.with.tags;tag1=value1;tag2=value2 9002 1534620625",
			name: "my_simple_metric_with_tags",
			sampleLabels: prometheus.Labels{
				"tag1": "value1",
				"tag2": "value2",
			},
			mappingPresent: false,
			value:          float64(9002),
		},
		{
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
		{
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
		{
			// new tags other than previously used, should drop
			line:     "my.simple.metric.with.tags;tag1=value1;tag3=value2 9002 1534620625",
			name:     "my_simple_metric_with_tags",
			willFail: true,
		},
	}

	c := newGraphiteCollector(log.NewNopLogger())

	for _, testCase := range testCases {

		if testCase.mappingPresent {
			c.mapper = &mockMapper{
				name:    testCase.name,
				labels:  testCase.mappingLabels,
				action:  testCase.action,
				present: testCase.mappingPresent,
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
	for _, k := range testCases {
		originalName := strings.Split(k.line, " ")[0]
		sample := c.samples[originalName]
		if k.willFail {
			assert.Nil(t, sample, "Found %s", k.name)
		} else {
			if assert.NotNil(t, sample, "Missing %s", k.name) {
				assert.Equal(t, k.name, sample.Name)
				assert.Equal(t, k.sampleLabels, sample.Labels)
				assert.Equal(t, k.value, sample.Value)
			}
		}
	}
}
