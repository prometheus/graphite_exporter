// Copyright 2014 The Prometheus Authors
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
	"io/ioutil"
	"regexp"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricLineRE      = regexp.MustCompile(`^(\*\.|[^*.]+\.|\*[^*.]+\.|[^*.]+\*\.|\.)*(\*|[^*.]+|\*[^*.]+|[^*.]+\*)$`)
	labelLineRE       = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*"(.*)"$`)
	invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9:_]`)
	validNameRE       = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9:_]*$`)
)

type metricMapping struct {
	regex  *regexp.Regexp
	labels prometheus.Labels
}

type metricMapper struct {
	mappings []metricMapping
	mutex    sync.RWMutex
}

type configLoadStates int

const (
	SEARCHING configLoadStates = iota
	METRIC_DEFINITION
)

func (m *metricMapper) initFromString(fileContents string) error {
	lines := strings.Split(fileContents, "\n")
	state := SEARCHING

	parsedMappings := []metricMapping{}
	currentMapping := metricMapping{labels: prometheus.Labels{}}
	for i, line := range lines {
		line := strings.TrimSpace(line)

		switch state {
		case SEARCHING:
			if line == "" {
				continue
			}

			if !metricLineRE.MatchString(line) {
				return fmt.Errorf("Line %d: expected metric match line, got: %s", i, line)
			}

			// Translate the glob-style metric match line into a proper regex that we
			// can use to match metrics later on.
			metricRe := strings.Replace(line, ".", "\\.", -1)
			metricRe = strings.Replace(metricRe, "*", "([^.]+)", -1)
			currentMapping.regex = regexp.MustCompile("^" + metricRe + "$")

			state = METRIC_DEFINITION

		case METRIC_DEFINITION:
			if line == "" {
				if len(currentMapping.labels) == 0 {
					return fmt.Errorf("Line %d: metric mapping didn't set any labels", i)
				}
				if _, ok := currentMapping.labels["name"]; !ok {
					return fmt.Errorf("Line %d: metric mapping didn't set a metric name", i)
				}

				parsedMappings = append(parsedMappings, currentMapping)

				state = SEARCHING
				currentMapping = metricMapping{labels: prometheus.Labels{}}
				continue
			}

			matches := labelLineRE.FindStringSubmatch(line)
			if len(matches) != 3 {
				return fmt.Errorf("Line %d: expected label mapping line, got: %s", i, line)
			}
			label, value := matches[1], matches[2]
			currentMapping.labels[label] = value
		default:
			panic("illegal state")
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.mappings = parsedMappings
	return nil
}

func (m *metricMapper) initFromFile(fileName string) error {
	mappingStr, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	return m.initFromString(string(mappingStr))
}

func (m *metricMapper) getMapping(metric string) (labels prometheus.Labels, present bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, mapping := range m.mappings {
		matches := mapping.regex.FindStringSubmatchIndex(metric)
		if len(matches) == 0 {
			continue
		}

		labels := prometheus.Labels{}
		for label, valueExpr := range mapping.labels {
			value := string(mapping.regex.ExpandString([]byte{}, valueExpr, metric, matches))
			if label == "name" {
				value = invalidNameCharRE.ReplaceAllString(value, "_")
				if !validNameRE.MatchString(value) {
					// Begins with a number or colon.
					value = "_" + value
				}
			}
			labels[label] = value
		}
		return labels, true
	}

	return nil, false
}
