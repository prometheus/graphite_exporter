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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
)

var (
	logger = log.NewNopLogger()
	c      = NewGraphiteCollector(logger, false, 5*time.Minute)

	now = time.Now()

	rawInput = `rspamd.actions.add_header 2 NOW
rspamd.actions;action=greylist 0 NOW
rspamd.actions;action=no_action 24 NOW
rspamd.actions;action=reject 1 NOW
rspamd.actions;action=rewrite_subject 0 NOW
rspamd.actions.soft_reject 0 NOW
rspamd.bytes_allocated 4165268944 NOW
rspamd.chunks_allocated 4294966730 NOW
rspamd.chunks_freed 0 NOW
rspamd.chunks_oversized 1 NOW
rspamd.connections 1 NOW
rspamd.control_connections 1 NOW
rspamd.ham_count 24 NOW
rspamd.learned 2 NOW
rspamd.pools_allocated 59 NOW
rspamd.pools_freed 171 NOW
rspamd.scanned 27 NOW
rspamd.shared_chunks_allocated 34 NOW
rspamd.spam_count 3 NOW`
	rawInput2 = strings.NewReplacer("NOW", fmt.Sprintf("%d", now.Unix())).Replace(rawInput)
	input     = strings.Split(rawInput2, "\n")

	// The name should be the same length to ensure the only difference is the tag parsing
	untaggedLine = fmt.Sprintf("rspamd.actions 2 %d", now.Unix())
	taggedLine   = fmt.Sprintf("rspamd.actions;action=add_header;foo=bar 2 %d", now.Unix())
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

func init() {
	c.mapper = &mockMapper{
		name:    "not_used",
		present: false,
	}
}

func benchmarkProcessLines(times int, b *testing.B, lines []string) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < times; i++ {
			for _, l := range lines {
				c.processLine(l)
			}
		}
	}
}

func benchmarkProcessLine(b *testing.B, line string) {
	// always report allocations since this is a hot path
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		c.processLine(line)
	}
}

// Mixed lines benchmarks
func BenchmarkProcessLineMixed1(b *testing.B) {
	benchmarkProcessLines(1, b, input)
}
func BenchmarkProcessLineMixed5(b *testing.B) {
	benchmarkProcessLines(5, b, input)
}
func BenchmarkProcessLineMixed50(b *testing.B) {
	benchmarkProcessLines(50, b, input)
}

// Individual line benchmarks
func BenchmarkProcessLineUntagged(b *testing.B) {
	benchmarkProcessLine(b, untaggedLine)
}
func BenchmarkProcessLineTagged(b *testing.B) {
	benchmarkProcessLine(b, taggedLine)
}
