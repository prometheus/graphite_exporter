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
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestParseNameAndTags(t *testing.T) {
	type testCase struct {
		line       string
		parsedName string
		labels     prometheus.Labels
		willFail   bool
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
			line:       "my_simple_metric_with_bad_tags;tag1=value1;tag2",
			parsedName: "my_simple_metric_with_bad_tags;tag1=value1;tag2",
			labels:     prometheus.Labels{},
			willFail:   true,
		},
	}

	tagErrorsTest := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphite_tag_parse_errors_test",
			Help: "Total count of samples with invalid tags",
		},
	)

	for _, testCase := range testCases {
		labels := prometheus.Labels{}
		n, err := parseMetricNameAndTags(testCase.line, labels, tagErrorsTest)

		if !testCase.willFail {
			assert.NoError(t, err, "Got unexpected error parsing %s", testCase.line)
		}

		assert.Equal(t, testCase.parsedName, n)
		assert.Equal(t, testCase.labels, labels)
	}
}
