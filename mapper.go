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
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricLineRE      = regexp.MustCompile(`^(\*\.|[^*.]+\.|\.)*(\*|[^*.]+)$`)
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

func (m *metricMapper) initFromString(fileContents string) error {
	parsedMappings := []metricMapping{}
	parser := newParser(fileContents)

	for {
		var mapping metricMapping
		err := parser.parseNext(&mapping)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		parsedMappings = append(parsedMappings, mapping)
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

type metricParser struct {
	lines []string
	index int
}

func newParser(fileContents string) *metricParser {
	return &metricParser{lines: strings.Split(fileContents, "\n")}
}

func isMetricLine(line string) bool {
	return isParsableLine(line) && metricLineRE.MatchString(line)
}

func isLabelLine(line string) bool {
	return isParsableLine(line) && labelLineRE.MatchString(line)
}

func isParsableLine(line string) bool {
	return line != "" && !strings.HasPrefix(line, "#")
}

func (r *metricParser) parseNext(mapping *metricMapping) error {
	searching := true
	var metricLine int

	for ; r.index < len(r.lines); r.index++ {
		lineText := strings.TrimSpace(r.lines[r.index])
		currentLine := r.index + 1

		if isLabelLine(lineText) {
			if searching {
				return fmt.Errorf("Line %d: expected metric match line, got: %s", currentLine, lineText)
			}

			matches := labelLineRE.FindStringSubmatch(lineText)
			if len(matches) != 3 {
				return fmt.Errorf("Line %d: expected label mapping line, got: %s", currentLine, lineText)
			}

			label, value := matches[1], matches[2]
			mapping.labels[label] = value
		} else if isMetricLine(lineText) {
			if !searching {
				break
			}

			searching = false
			metricLine = currentLine

			metricRe := strings.Replace(lineText, ".", "\\.", -1)
			metricRe = strings.Replace(metricRe, "*", "([^.]+)", -1)
			mapping.regex = regexp.MustCompile("^" + metricRe + "$")
			mapping.labels = prometheus.Labels{}
		}
	}

	if !searching {
		if len(mapping.labels) == 0 {
			return fmt.Errorf("Line %d: metric mapping didn't set any labels", metricLine)
		}

		if _, ok := mapping.labels["name"]; !ok {
			return fmt.Errorf("Line %d: metric mapping didn't set a metric name", metricLine)
		}

		return nil
	}

	return io.EOF
}
