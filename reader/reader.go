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
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-graphite/go-whisper"
)

type DBReader interface {
	Metrics() ([]string, error)
	GetMinAndMaxTimestamps() (int64, int64, error)
	Points(string, int64, int64) ([]Point, error)
}

type Point struct {
	Timestamp int64
	Value     float64
}

func NewReader(path string) DBReader {
	return &whisperReader{
		path: path,
	}
}

type whisperReader struct {
	path string
	wdb  whisper.Whisper
}

func (w *whisperReader) Metrics() ([]string, error) {
	metrics := make([]string, 0)
	err := filepath.Walk(w.path, func(path string, _ os.FileInfo, _ error) error {
		if !strings.HasSuffix(path, ".wsp") {
			return nil
		}
		path = strings.TrimPrefix(path, strings.TrimSuffix(w.path, string(os.PathSeparator))+string(os.PathSeparator))
		path = strings.TrimSuffix(path, ".wsp")
		metrics = append(metrics, strings.ReplaceAll(path, string(os.PathSeparator), "."))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (w *whisperReader) opendb(metric string) (*whisper.Whisper, error) {
	path := path.Join(append([]string{w.path}, strings.Split(metric, ".")...)...) + ".wsp"
	flag := os.O_RDONLY
	return whisper.OpenWithOptions(path, &whisper.Options{
		FLock:        false,
		OpenFileFlag: &flag,
	})
}

func (w *whisperReader) GetMinAndMaxTimestamps() (int64, int64, error) {
	var (
		// Go-Graphite timestamps are int32.
		sampleMin = math.MaxInt32
		sampleMax = math.MinInt32
	)
	metrics, err := w.Metrics()
	if err != nil {
		return 0, 0, err
	}
	for _, metric := range metrics {
		wdb, err := w.opendb(metric)
		if err != nil {
			return 0, 0, err
		}
		ts, err := wdb.Fetch(math.MinInt32, math.MaxInt32)
		if err != nil {
			return 0, 0, err
		}
		for _, sample := range ts.Points() {
			if math.IsNaN(sample.Value) {
				continue
			}
			if sample.Time < sampleMin {
				sampleMin = sample.Time
			}
			if sample.Time > sampleMax {
				sampleMax = sample.Time
			}
		}
		err = wdb.Close()
		if err != nil {
			return 0, 0, err
		}
	}
	if sampleMin > sampleMax {
		return 0, 0, fmt.Errorf("no valid sample found (min: %d, max: %v)", sampleMin, metrics)
	}
	return int64(1000 * sampleMin), int64(1000 * sampleMax), nil
}

func (w *whisperReader) Points(metric string, from, until int64) ([]Point, error) {
	wdb, err := w.opendb(metric)
	if err != nil {
		return nil, err
	}
	defer wdb.Close()
	ts, err := wdb.Fetch(int(from/1000), int(until/1000))
	if err != nil {
		return nil, err
	}
	points := make([]Point, 0)
	for _, sample := range ts.Points() {
		if math.IsNaN(sample.Value) {
			continue
		}
		points = append(points, Point{Timestamp: 1000 * int64(sample.Time), Value: sample.Value})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp < points[j].Timestamp })
	return points, nil
}
