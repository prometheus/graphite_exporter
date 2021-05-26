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

// +build !aix

package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/units"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/graphite_exporter/reader"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/tsdb"
	tsdb_errors "github.com/prometheus/prometheus/tsdb/errors"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
)

var invalidMetricChars = regexp.MustCompile("[^a-zA-Z0-9_:]")

func createBlocks(input reader.DBReader, mint, maxt, blockDuration int64, maxSamplesInAppender int, outputDir string, metricMapper *mapper.MetricMapper, strictMatch, humanReadable bool) (returnErr error) {
	mint = blockDuration * (mint / blockDuration)

	db, err := tsdb.OpenDBReadOnly(outputDir, nil)
	if err != nil {
		return err
	}
	defer func() {
		returnErr = tsdb_errors.NewMulti(returnErr, db.Close()).Err()
	}()

	var wroteHeader bool

	metrics, err := input.Metrics()
	if err != nil {
		return err
	}

	for t := mint; t <= maxt; t = t + blockDuration {
		tsUpper := t + blockDuration
		err := func() error {
			// To prevent races with compaction, a block writer only allows appending samples
			// that are at most half a block size older than the most recent sample appended so far.
			// However, in the way we use the block writer here, compaction doesn't happen, while we
			// also need to append samples throughout the whole block range. To allow that, we
			// pretend that the block is twice as large here, but only really add sample in the
			// original interval later.
			w, err := tsdb.NewBlockWriter(log.NewNopLogger(), outputDir, 2*blockDuration)
			if err != nil {
				return errors.Wrap(err, "block writer")
			}
			defer func() {
				err = tsdb_errors.NewMulti(err, w.Close()).Err()
			}()

			ctx := context.Background()
			app := w.Appender(ctx)
			samplesCount := 0
			for _, m := range metrics {
				mapping, mappingLabels, mappingPresent := metricMapper.GetMapping(m, mapper.MetricTypeGauge)

				if (mappingPresent && mapping.Action == mapper.ActionTypeDrop) || (!mappingPresent && strictMatch) {
					continue
				}

				l := make(labels.Labels, 0)
				// add mapping labels to parsed labels·
				for k, v := range mappingLabels {
					l = append(l, labels.Label{Name: k, Value: v})
				}

				var name string
				if mappingPresent {
					name = invalidMetricChars.ReplaceAllString(mapping.Name, "_")
				} else {
					name = invalidMetricChars.ReplaceAllString(m, "_")
				}
				l = append(l, labels.Label{Name: "__name__", Value: name})

				points, err := input.Points(m, t, tsUpper)
				if err != nil {
					return err
				}
				for _, point := range points {
					if _, err := app.Add(l, point.Timestamp, point.Value); err != nil {
						return errors.Wrap(err, "add sample")
					}

					samplesCount++
					if samplesCount < maxSamplesInAppender {
						continue
					}

					// If we arrive here, the samples count is greater than the maxSamplesInAppender.
					// Therefore the old appender is committed and a new one is created.
					// This prevents keeping too many samples lined up in an appender and thus in RAM.
					if err := app.Commit(); err != nil {
						return errors.Wrap(err, "commit")
					}

					app = w.Appender(ctx)
					samplesCount = 0
				}
			}

			if err := app.Commit(); err != nil {
				return errors.Wrap(err, "commit")
			}

			block, err := w.Flush(ctx)
			switch err {
			case nil:
				blocks, err := db.Blocks()
				if err != nil {
					return errors.Wrap(err, "get blocks")
				}
				for _, b := range blocks {
					if b.Meta().ULID == block {
						printBlocks([]tsdb.BlockReader{b}, !wroteHeader, humanReadable)
						wroteHeader = true
						break
					}
				}
			case tsdb.ErrNoSeriesAppended:
			default:
				return errors.Wrap(err, "flush")
			}

			return nil
		}()

		if err != nil {
			return errors.Wrap(err, "process blocks")
		}
	}
	return nil

}

func printBlocks(blocks []tsdb.BlockReader, writeHeader, humanReadable bool) {
	tw := tabwriter.NewWriter(os.Stdout, 13, 0, 2, ' ', 0)
	defer tw.Flush()

	if writeHeader {
		fmt.Fprintln(tw, "BLOCK ULID\tMIN TIME\tMAX TIME\tDURATION\tNUM SAMPLES\tNUM CHUNKS\tNUM SERIES\tSIZE")
	}

	for _, b := range blocks {
		meta := b.Meta()

		fmt.Fprintf(tw,
			"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			meta.ULID,
			getFormatedTime(meta.MinTime, humanReadable),
			getFormatedTime(meta.MaxTime, humanReadable),
			time.Duration(meta.MaxTime-meta.MinTime)*time.Millisecond,
			meta.Stats.NumSamples,
			meta.Stats.NumChunks,
			meta.Stats.NumSeries,
			getFormatedBytes(b.Size(), humanReadable),
		)
	}
}

func checkErr(err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func getFormatedTime(timestamp int64, humanReadable bool) string {
	if humanReadable {
		return time.Unix(timestamp/1000, 0).UTC().String()
	}
	return strconv.FormatInt(timestamp, 10)
}

func getFormatedBytes(bytes int64, humanReadable bool) string {
	if humanReadable {
		return units.Base2Bytes(bytes).String()
	}
	return strconv.FormatInt(bytes, 10)
}

func backfill(maxSamplesInAppender int, inputDir, outputDir, mappingConfig string, strictMatch, humanReadable bool, blockDuration int64) (err error) {
	var (
		// Those do not really matter when backfilling.
		cacheOption = mapper.WithCacheType("lru")
		cacheSize   = 1
	)

	wdb := reader.NewReader(inputDir)
	mint, maxt, err := wdb.GetMinAndMaxTimestamps()
	if err != nil {
		return errors.Wrap(err, "getting min and max timestamp")
	}
	metricMapper := &mapper.MetricMapper{}

	if mappingConfig != "" {
		err := metricMapper.InitFromFile(mappingConfig, cacheSize, cacheOption)
		if err != nil {
			logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
			level.Error(logger).Log("msg", "Error loading metric mapping config", "err", err)
			return err
		}
	} else {
		metricMapper.InitCache(cacheSize, cacheOption)
	}

	return errors.Wrap(createBlocks(wdb, mint, maxt, blockDuration, maxSamplesInAppender, outputDir, metricMapper, strictMatch, humanReadable), "block creation")
}

func backfillWhisper(inputDir, outputDir, mappingConfig string, strictMatch, humanReadable bool, optBlockDuration time.Duration) (err error) {
	blockDuration := int64(time.Duration(optBlockDuration) / time.Millisecond)

	if !validateBlockDuration(blockDuration) {
		return fmt.Errorf("invalid block duration: %s", optBlockDuration.String())
	}

	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return errors.Wrap(err, "create output dir")
	}

	return backfill(5000, inputDir, outputDir, mappingConfig, strictMatch, humanReadable, blockDuration)
}

func validateBlockDuration(t int64) bool {
	i, f := math.Modf(float64(t) / float64(tsdb.DefaultBlockDuration))
	if f != 0 {
		return false
	}

	i, f = math.Modf(math.Log2(i))
	return f == 0 && i >= 0
}
