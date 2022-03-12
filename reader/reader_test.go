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

package reader

import (
	"math"
	"sort"
	"testing"
	"time"

	"github.com/go-graphite/go-whisper"
	"github.com/stretchr/testify/require"
)

func init() {
	// Pin go-whisper to a fixed timestamp so that the test data is in the window of retention.
	whisper.Now = func() time.Time { return time.Unix(1640000000, 0) }
}

func TestListMetrics(t *testing.T) {
	reader := NewReader("testdata")
	metrics, err := reader.Metrics()
	require.NoError(t, err)

	sort.Strings(metrics)
	expected := []string{
		"test-whisper.load.load.longterm",
		"test-whisper.load.load.shortterm",
		"test-whisper.load.load.midterm",
	}

	sort.Strings(expected)

	require.Equal(t, expected, metrics)
}

func TestGetMinAndMaxTimestamp(t *testing.T) {
	reader := NewReader("testdata")
	min, max, err := reader.GetMinAndMaxTimestamps()
	require.NoError(t, err)

	require.True(t, min > math.MinInt64)
	require.Equal(t, int64(1611068400000), max)
}

func TestGetPoints(t *testing.T) {
	reader := NewReader("testdata")
	points, err := reader.Points("test-whisper.load.load.longterm", 1000*math.MinInt32, 1000*math.MaxInt32)
	require.NoError(t, err)

	expectedLastPoints := []Point{
		{
			Timestamp: 1611067800000,
			Value:     1.0511666666666666,
		},
		{
			Timestamp: 1611068400000,
			Value:     1.0636666666666668,
		},
	}

	for i, p := range expectedLastPoints {
		pos := len(points) - len(expectedLastPoints) + i
		require.Equal(t, p, points[pos])
	}
}
