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

//go:build !aix
// +build !aix

package main

import (
	"context"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-graphite/go-whisper"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/stretchr/testify/require"
)

func TestBackfill(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	for _, tt := range []struct {
		name          string
		metricName    string
		labels        map[string]string
		mappingConfig string
		strictMatch   bool
	}{
		{
			name:       "default",
			metricName: "load_cpu_cpu0",
		},
		{
			name: "with_mapping",
			mappingConfig: `
mappings:
- match: "load.*.*"
  name: load_$1
  labels:
    state: idle
    cpu: $2`,
			metricName: "load_cpu",
			labels:     map[string]string{"cpu": "cpu0", "state": "idle"},
		},
		{
			name:        "strict_match",
			strictMatch: true,
			mappingConfig: `
mappings:
- match: load.*.*
  name: load_$1
  labels:
    cpu: $2`,
			metricName: "load_cpu",
			labels:     map[string]string{"cpu": "cpu0"},
		},
	} {
		tt := tt // TODO(matthias): remove after upgrading to Go 1.22
		t.Run(tt.name, func(t *testing.T) {

			var (
				metricTime = int(time.Now().Add(-30 * time.Minute).Unix())
				tmpData    = filepath.Join(os.TempDir(), "graphite_exporter_test")
				whisperDir = filepath.Join(tmpData, "whisper", "load", "cpu")
			)

			defer os.RemoveAll(tmpData)

			require.NoError(t, os.MkdirAll(whisperDir, 0o777))
			retentions, err := whisper.ParseRetentionDefs("1s:3600")
			require.NoError(t, err)
			wsp, err := whisper.Create(filepath.Join(whisperDir, "cpu0.wsp"), retentions, whisper.Sum, 0.5)
			require.NoError(t, err)
			require.NoError(t, wsp.Update(1234.5678, metricTime-1))
			require.NoError(t, wsp.Update(12345.678, metricTime))
			require.NoError(t, wsp.Close())

			arguments := []string{
				"-test.main",
				"create-blocks",
			}

			if tt.mappingConfig != "" {
				cfgFile := filepath.Join(tmpData, "mapping.yaml")
				err := os.WriteFile(cfgFile, []byte(tt.mappingConfig), 0644)
				require.NoError(t, err)
				arguments = append(arguments, "--graphite.mapping-config", cfgFile)
			}

			if tt.strictMatch {
				arguments = append(arguments, "--graphite.mapping-strict-match")
			}

			arguments = append(arguments, filepath.Join(tmpData, "whisper"), filepath.Join(tmpData, "data"))

			cmd := exec.Command(testPath, arguments...)

			// Log stderr in case of failure.
			stderr, err := cmd.StderrPipe()
			require.NoError(t, err)
			go func() {
				slurp, _ := io.ReadAll(stderr)
				t.Log(string(slurp))
			}()

			err = cmd.Start()
			require.NoError(t, err)

			err = cmd.Wait()
			require.NoError(t, err)

			require.NoError(t, os.MkdirAll(filepath.Join(tmpData, "data", "wal"), 0o777))

			db, err := tsdb.OpenDBReadOnly(filepath.Join(tmpData, "data"), "", nil)
			require.NoError(t, err)
			q, err := db.Querier(math.MinInt64, math.MaxInt64)
			require.NoError(t, err)

			s := queryAllSeries(t, q)

			ll := labels.FromMap(tt.labels)

			//Prepend the label __name__ to match expected order
			ll = append([]labels.Label{
				{
					Name:  "__name__",
					Value: tt.metricName,
				},
			}, ll...)

			require.Equal(t, ll, s[0].Labels)
			require.Equal(t, 1000*int64(metricTime-1), s[0].Timestamp)
			require.Equal(t, 1234.5678, s[0].Value)
			require.Equal(t, ll, s[1].Labels)
			require.Equal(t, 1000*int64(metricTime), s[1].Timestamp)
			require.Equal(t, 12345.678, s[1].Value)
		})
	}
}

type backfillSample struct {
	Timestamp int64
	Value     float64
	Labels    labels.Labels
}

func queryAllSeries(t *testing.T, q storage.Querier) []backfillSample {
	ss := q.Select(context.Background(), false, nil, labels.MustNewMatcher(labels.MatchRegexp, "", ".*"))
	samples := []backfillSample{}
	for ss.Next() {
		series := ss.At()
		it := series.Iterator(nil)
		require.NoError(t, it.Err())
		for it.Next() != chunkenc.ValNone {
			ts, v := it.At()
			samples = append(samples, backfillSample{Timestamp: ts, Value: v, Labels: series.Labels()})
		}
	}
	return samples
}

func TestValidateBlockSize(t *testing.T) {
	require.True(t, validateBlockDuration(int64(time.Duration(2*time.Hour)/time.Millisecond)))
	require.True(t, validateBlockDuration(int64(time.Duration(4*time.Hour)/time.Millisecond)))
	require.True(t, validateBlockDuration(int64(time.Duration(16*time.Hour)/time.Millisecond)))

	require.False(t, validateBlockDuration(0))
	require.False(t, validateBlockDuration(int64(time.Duration(1*time.Hour)/time.Millisecond)))
	require.False(t, validateBlockDuration(int64(time.Duration(3*time.Hour)/time.Millisecond)))
	require.False(t, validateBlockDuration(int64(time.Duration(11*time.Hour)/time.Millisecond)))
}
