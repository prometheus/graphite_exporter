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
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-graphite/go-whisper"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb"
	"github.com/stretchr/testify/require"
)

func TestBackfill(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	var (
		metricTime = int(time.Now().Add(-30 * time.Minute).Unix())
		whisperDir = filepath.Join(tmpData, "whisper", "load", "cpu")
	)

	require.NoError(t, os.MkdirAll(whisperDir, 0777))
	retentions, err := whisper.ParseRetentionDefs("1s:3600")
	require.NoError(t, err)
	wsp, err := whisper.Create(filepath.Join(whisperDir, "cpu0.wsp"), retentions, whisper.Sum, 0.5)
	require.NoError(t, err)
	require.NoError(t, wsp.Update(1234.5678, metricTime-1))
	require.NoError(t, wsp.Update(12345.678, metricTime))
	require.NoError(t, wsp.Close())

	cmd := exec.Command(testPath, "-test.main", "create-blocks", filepath.Join(tmpData, "whisper"), filepath.Join(tmpData, "data"))

	// Log stderr in case of failure.
	stderr, err := cmd.StderrPipe()
	require.NoError(t, err)
	go func() {
		slurp, _ := ioutil.ReadAll(stderr)
		t.Log(string(slurp))
	}()

	err = cmd.Start()
	require.NoError(t, err)

	err = cmd.Wait()
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(tmpData, "data", "wal"), 0777))

	db, err := tsdb.OpenDBReadOnly(filepath.Join(tmpData, "data"), nil)
	require.NoError(t, err)
	q, err := db.Querier(context.TODO(), math.MinInt64, math.MaxInt64)
	require.NoError(t, err)

	s := queryAllSeries(t, q)

	require.Equal(t, labels.FromStrings("__name__", "load_cpu_cpu0"), s[0].Labels)
	require.Equal(t, 1000*int64(metricTime-1), s[0].Timestamp)
	require.Equal(t, 1234.5678, s[0].Value)
	require.Equal(t, labels.FromStrings("__name__", "load_cpu_cpu0"), s[1].Labels)
	require.Equal(t, 1000*int64(metricTime), s[1].Timestamp)
	require.Equal(t, 12345.678, s[1].Value)
}

type backfillSample struct {
	Timestamp int64
	Value     float64
	Labels    labels.Labels
}

func queryAllSeries(t *testing.T, q storage.Querier) []backfillSample {
	ss := q.Select(false, nil, labels.MustNewMatcher(labels.MatchRegexp, "", ".*"))
	samples := []backfillSample{}
	for ss.Next() {
		series := ss.At()
		it := series.Iterator()
		require.NoError(t, it.Err())
		for it.Next() {
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
