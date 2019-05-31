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
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

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

func (m *mockMapper) InitFromFile(string) error {
	return nil
}

func TestProcessLine(t *testing.T) {

	type testCase struct {
		line     string
		name     string
		labels   map[string]string
		value    float64
		present  bool
		willFail bool
		action   mapper.ActionType
		strict   bool
	}

	testCases := []testCase{
		{
			line: "my.simple.metric 9001 1534620625",
			name: "my_simple_metric",
			labels: map[string]string{
				"foo":  "bar",
				"zip":  "zot",
				"name": "alabel",
			},
			present: true,
			value:   float64(9001),
		},
		{
			line: "my.simple.metric.baz 9002 1534620625",
			name: "my_simple_metric",
			labels: map[string]string{
				"baz": "bat",
			},
			present: true,
			value:   float64(9002),
		},
		{
			line:    "my.nomap.metric 9001 1534620625",
			name:    "my_nomap_metric",
			value:   float64(9001),
			present: false,
		},
		{
			line:     "my.nomap.metric.novalue 9001 ",
			name:     "my_nomap_metric_novalue",
			labels:   nil,
			value:    float64(9001),
			willFail: true,
		},
		{
			line:     "my.mapped.metric.drop 55 1534620625",
			name:     "my_mapped_metric_drop",
			present:  true,
			willFail: true,
			action:   mapper.ActionTypeDrop,
		},
		{
			line:     "my.mapped.strict.metric 55 1534620625",
			name:     "my_mapped_strict_metric",
			value:    float64(55),
			present:  true,
			willFail: false,
			strict:   true,
		},
		{
			line:     "my.mapped.strict.metric.drop 55 1534620625",
			name:     "my_mapped_strict_metric_drop",
			present:  false,
			willFail: true,
			strict:   true,
		},
	}

	c := newGraphiteCollector(&mockMapper{})
	defer c.stop()

	for _, testCase := range testCases {

		if testCase.present {
			c.mapper = &mockMapper{
				name:    testCase.name,
				labels:  testCase.labels,
				action:  testCase.action,
				present: testCase.present,
			}
		} else {
			c.mapper = &mockMapper{
				present: testCase.present,
			}
		}

		c.strictMatch = testCase.strict
		c.processLine(testCase.line)

	}

	c.ch <- nil
	for _, k := range testCases {
		originalName := strings.Split(k.line, " ")[0]
		sample := c.samples[originalName]
		if k.willFail {
			assert.Nil(t, sample, "Found %s", k.name)
		} else {
			if assert.NotNil(t, sample, "Missing %s", k.name) {
				assert.Equal(t, k.name, sample.Name)
				assert.Equal(t, k.labels, sample.Labels)
				assert.Equal(t, k.value, sample.Value)
			}
		}
	}
}

func newCollectorForTest(t *testing.T, m metricMapper) (*graphiteCollector, prometheus.Gatherer) {
	collector := newGraphiteCollector(&mapper.MetricMapper{})

	registry := prometheus.NewRegistry()
	if err := registry.Register(collector); err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	return collector, registry
}

func testGathering(t *testing.T, mapper metricMapper, send func(c *graphiteCollector)) {
	c, registry := newCollectorForTest(t, mapper)
	defer c.stop()

	send(c)

	// TODO: check the results
	if _, err := registry.Gather(); err != nil {
		t.Errorf("failed to gather metrics: %v", err)
	}
}

// TestGathering checks whether we can successfully gather metrics after
// sending some samples. Much of the validation happens at scrape time, and we
// frequently run into issues that only manifest there.
func TestGathering(t *testing.T) {
	metricLine := fmt.Sprintf("my.metric 42 %d", time.Now().Unix()-2)

	t.Run("single", func(t *testing.T) {
		testGathering(
			t,
			&mapper.MetricMapper{},
			func(c *graphiteCollector) { c.processReader(strings.NewReader(metricLine)) },
		)
	})

	t.Run("serial", func(t *testing.T) {
		testGathering(
			t,
			&mapper.MetricMapper{},
			func(c *graphiteCollector) {
				for i := 0; i < 1000; i++ {
					c.processReader(strings.NewReader(metricLine))
				}
			},
		)
	})

	t.Run("parallel", func(t *testing.T) {
		testGathering(
			t,
			&mapper.MetricMapper{},
			func(c *graphiteCollector) {
				w := sync.WaitGroup{}
				for i := 0; i < 1000; i++ {
					w.Add(1)
					go func() {
						c.processReader(strings.NewReader(metricLine))
						w.Done()
					}()
				}
				w.Wait()
			},
		)
	})

	t.Run("issue90_serial", func(t *testing.T) {
		m := &mapper.MetricMapper{}
		if err := m.InitFromFile("e2e/fixtures/issue90.yml"); err != nil {
			t.Fatalf("failed to initialize mapper: %v", err
		}

		testGathering(
			t,
			m,
			func(c *graphiteCollector) {
				f, err := os.Open("e2e/fixtures/issue90_in.txt")
				if err != nil {
					t.Fatalf("failed to read fixture: %v", err)
				}
				defer f.Close()
				c.processReader(f)
			},
		)
	})
}
