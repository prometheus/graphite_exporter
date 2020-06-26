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

package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func benchmarkProcessLine(times int, b *testing.B) {
	logger := log.NewNopLogger()
	c := newGraphiteCollector(logger)

	now := time.Now()

	rawInput := `rspamd.actions.add_header 2 NOW
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
	rawInput = strings.NewReplacer("NOW", fmt.Sprintf("%d", now.Unix())).Replace(rawInput)
	input := strings.Split(rawInput, "\n")

	c.mapper = &mockMapper{
		name:    "not_used",
		present: false,
	}

	// reset benchmark timer to not measure startup costs
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < times; i++ {
			for _, l := range input {
				c.processLine(l)
			}
		}
	}
}

func BenchmarkProcessLine1(b *testing.B) {
	benchmarkProcessLine(1, b)
}
func BenchmarkProcessLine5(b *testing.B) {
	benchmarkProcessLine(5, b)
}
func BenchmarkProcessLine50(b *testing.B) {
	benchmarkProcessLine(50, b)
}
