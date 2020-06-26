// Copyright 2015 The Prometheus Authors
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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress   = kingpin.Flag("web.listen-address", "Address on which to expose metrics.").Default(":9108").String()
	metricsPath     = kingpin.Flag("web.telemetry-path", "Path under which to expose Prometheus metrics.").Default("/metrics").String()
	graphiteAddress = kingpin.Flag("graphite.listen-address", "TCP and UDP address on which to accept samples.").Default(":9109").String()
	mappingConfig   = kingpin.Flag("graphite.mapping-config", "Metric mapping configuration file name.").Default("").String()
	sampleExpiry    = kingpin.Flag("graphite.sample-expiry", "How long a sample is valid for.").Default("5m").Duration()
	strictMatch     = kingpin.Flag("graphite.mapping-strict-match", "Only store metrics that match the mapping configuration.").Bool()
	cacheSize       = kingpin.Flag("graphite.cache-size", "Maximum size of your metric mapping cache. Relies on least recently used replacement policy if max size is reached.").Default("1000").Int()
	cacheType       = kingpin.Flag("graphite.cache-type", "Metric mapping cache type. Valid options are \"lru\" and \"random\"").Default("lru").Enum("lru", "random")
	dumpFSMPath     = kingpin.Flag("debug.dump-fsm", "The path to dump internal FSM generated for glob matching as Dot file.").Default("").String()

	lastProcessed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "graphite_last_processed_timestamp_seconds",
			Help: "Unix timestamp of the last processed graphite metric.",
		},
	)
	sampleExpiryMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "graphite_sample_expiry_seconds",
			Help: "How long in seconds a metric sample is valid for.",
		},
	)
	tagParseFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphite_tag_parse_failures",
			Help: "Total count of samples with invalid tags",
		})
	invalidMetrics = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphite_invalid_metrics",
			Help: "Total count of metrics dropped due to mismatched label keys",
		})
	invalidMetricChars  = regexp.MustCompile("[^a-zA-Z0-9_:]")
	metricNameKeysIndex = newMetricNameAndKeys()
)

type graphiteSample struct {
	OriginalName string
	Name         string
	Labels       prometheus.Labels
	Help         string
	Value        float64
	Type         prometheus.ValueType
	Timestamp    time.Time
}

func (s graphiteSample) String() string {
	return fmt.Sprintf("%#v", s)
}

type metricMapper interface {
	GetMapping(string, mapper.MetricType) (*mapper.MetricMapping, prometheus.Labels, bool)
	InitFromFile(string, int, ...mapper.CacheOption) error
	InitCache(int, ...mapper.CacheOption)
}

// metricNameAndKeys is a cache of metric names and the label keys previously used
type metricNameAndKeys struct {
	mtx   sync.Mutex
	cache map[string]string
}

func newMetricNameAndKeys() *metricNameAndKeys {
	x := metricNameAndKeys{
		cache: make(map[string]string),
	}
	return &x
}

func keysFromLabels(labels prometheus.Labels) string {
	labelKeys := make([]string, len(labels))
	for k, _ := range labels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)
	return strings.Join(labelKeys, ",")
}

// checkNameAndKeys returns true if metric has the same label keys or is new, false if not
func (c *metricNameAndKeys) checkNameAndKeys(name string, labels prometheus.Labels) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	providedKeys := keysFromLabels(labels)
	if keys, found := c.cache[name]; found {
		return keys == providedKeys
	}

	c.cache[name] = providedKeys
	return true
}

type graphiteCollector struct {
	samples     map[string]*graphiteSample
	mu          *sync.Mutex
	mapper      metricMapper
	sampleCh    chan *graphiteSample
	lineCh      chan string
	strictMatch bool
	logger      log.Logger
}

func newGraphiteCollector(logger log.Logger) *graphiteCollector {
	c := &graphiteCollector{
		sampleCh:    make(chan *graphiteSample),
		lineCh:      make(chan string),
		mu:          &sync.Mutex{},
		samples:     map[string]*graphiteSample{},
		strictMatch: *strictMatch,
		logger:      logger,
	}
	go c.processSamples()
	go c.processLines()
	return c
}

func (c *graphiteCollector) processReader(reader io.Reader) {
	lineScanner := bufio.NewScanner(reader)
	for {
		if ok := lineScanner.Scan(); !ok {
			break
		}
		c.lineCh <- lineScanner.Text()
	}
}

func (c *graphiteCollector) processLines() {
	for line := range c.lineCh {
		c.processLine(line)
	}
}

func parseMetricNameAndTags(name string, labels prometheus.Labels) (string, error) {
	if strings.ContainsRune(name, ';') {
		// name contains tags - parse tags and add to labels
		if strings.Count(name, ";") != strings.Count(name, "=") {
			tagParseFailures.Inc()
			return name, fmt.Errorf("error parsing tags on %s", name)
		}

		parts := strings.Split(name, ";")
		parsedName := parts[0]
		tags := parts[1:]

		for _, tag := range tags {
			kv := strings.SplitN(tag, "=", 2)
			if len(kv) != 2 {
				// we may have added bad labels already...
				tagParseFailures.Inc()
				return name, fmt.Errorf("error parsing tags on %s", name)
			}

			k := kv[0]
			v := kv[1]
			labels[k] = v
		}

		return parsedName, nil
	}

	return name, nil
}

func (c *graphiteCollector) processLine(line string) {
	line = strings.TrimSpace(line)
	level.Debug(c.logger).Log("msg", "Incoming line", "line", line)
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		level.Info(c.logger).Log("msg", "Invalid part count", "parts", len(parts), "line", line)
		return
	}
	originalName := parts[0]
	var name string
	var err error
	mapping, labels, mappingPresent := c.mapper.GetMapping(originalName, mapper.MetricTypeGauge)

	if (mappingPresent && mapping.Action == mapper.ActionTypeDrop) || (!mappingPresent && c.strictMatch) {
		return
	}

	if mappingPresent {
		parsedLabels := make(prometheus.Labels)
		_, err = parseMetricNameAndTags(originalName, parsedLabels)
		if err != nil {
			level.Info(c.logger).Log("msg", "Invalid tags", "line", line)
			return
		}

		name = invalidMetricChars.ReplaceAllString(mapping.Name, "_")
		// check to ensure the same tags are present
		if validKeys := metricNameKeysIndex.checkNameAndKeys(name, parsedLabels); !validKeys {
			level.Info(c.logger).Log("msg", "Dropped because metric keys do not match previously used keys", "line", line)
			invalidMetrics.Inc()
			return
		}
	} else {
		labels = make(prometheus.Labels)
		name, err = parseMetricNameAndTags(originalName, labels)
		if err != nil {
			level.Info(c.logger).Log("msg", "Invalid tags", "line", line)
			return
		}
		name = invalidMetricChars.ReplaceAllString(name, "_")
		// check to ensure the same tags are present
		if validKeys := metricNameKeysIndex.checkNameAndKeys(name, labels); !validKeys {
			level.Info(c.logger).Log("msg", "Dropped because metric keys do not match previously used keys", "line", line)
			invalidMetrics.Inc()
			return
		}
	}

	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		level.Info(c.logger).Log("msg", "Invalid value", "line", line)
		return
	}
	timestamp, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		level.Info(c.logger).Log("msg", "Invalid timestamp", "line", line)
		return
	}
	sample := graphiteSample{
		OriginalName: originalName,
		Name:         name,
		Value:        value,
		Labels:       labels,
		Type:         prometheus.GaugeValue,
		Help:         fmt.Sprintf("Graphite metric %s", name),
		Timestamp:    time.Unix(int64(timestamp), int64(math.Mod(timestamp, 1.0)*1e9)),
	}
	level.Debug(c.logger).Log("msg", "Processing sample", "sample", sample)
	lastProcessed.Set(float64(time.Now().UnixNano()) / 1e9)
	c.sampleCh <- &sample
}

func (c *graphiteCollector) processSamples() {
	ticker := time.NewTicker(time.Minute).C

	for {
		select {
		case sample, ok := <-c.sampleCh:
			if sample == nil || !ok {
				return
			}
			c.mu.Lock()
			c.samples[sample.OriginalName] = sample
			c.mu.Unlock()
		case <-ticker:
			// Garbage collect expired samples.
			ageLimit := time.Now().Add(-*sampleExpiry)
			c.mu.Lock()
			for k, sample := range c.samples {
				if ageLimit.After(sample.Timestamp) {
					delete(c.samples, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Collect implements prometheus.Collector.
func (c graphiteCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- lastProcessed

	c.mu.Lock()
	samples := make([]*graphiteSample, 0, len(c.samples))
	for _, sample := range c.samples {
		samples = append(samples, sample)
	}
	c.mu.Unlock()

	ageLimit := time.Now().Add(-*sampleExpiry)
	for _, sample := range samples {
		if ageLimit.After(sample.Timestamp) {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(sample.Name, sample.Help, []string{}, sample.Labels),
			sample.Type,
			sample.Value,
		)
	}
}

// Describe implements prometheus.Collector.
func (c graphiteCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lastProcessed.Desc()
}

func init() {
	prometheus.MustRegister(version.NewCollector("graphite_exporter"))
}

func dumpFSM(mapper *mapper.MetricMapper, dumpFilename string, logger log.Logger) error {
	if mapper.FSM == nil {
		return fmt.Errorf("no FSM available to be dumped, possibly because the mapping contains regex patterns")
	}
	f, err := os.Create(dumpFilename)
	if err != nil {
		return err
	}
	level.Info(logger).Log("msg", "Start dumping FSM", "to", dumpFilename)
	w := bufio.NewWriter(f)
	mapper.FSM.DumpFSM(w)
	w.Flush()
	f.Close()
	level.Info(logger).Log("msg", "Finish dumping FSM")
	return nil
}

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("graphite_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	prometheus.MustRegister(sampleExpiryMetric)
	prometheus.MustRegister(tagParseFailures)
	prometheus.MustRegister(invalidMetrics)
	sampleExpiryMetric.Set(sampleExpiry.Seconds())

	level.Info(logger).Log("msg", "Starting graphite_exporter", "version_info", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	c := newGraphiteCollector(logger)
	prometheus.MustRegister(c)

	c.mapper = &mapper.MetricMapper{}
	cacheOption := mapper.WithCacheType(*cacheType)

	if *mappingConfig != "" {
		err := c.mapper.InitFromFile(*mappingConfig, *cacheSize, cacheOption)
		if err != nil {
			level.Error(logger).Log("msg", "Error loading metric mapping config", "err", err)
			os.Exit(1)
		}
	} else {
		c.mapper.InitCache(*cacheSize, cacheOption)
	}

	if *dumpFSMPath != "" {
		err := dumpFSM(c.mapper.(*mapper.MetricMapper), *dumpFSMPath, logger)
		if err != nil {
			level.Error(logger).Log("msg", "Error dumping FSM", "err", err)
			os.Exit(1)
		}
	}

	tcpSock, err := net.Listen("tcp", *graphiteAddress)
	if err != nil {
		level.Error(logger).Log("msg", "Error binding to TCP socket", "err", err)
		os.Exit(1)
	}
	go func() {
		for {
			conn, err := tcpSock.Accept()
			if err != nil {
				level.Error(logger).Log("msg", "Error accepting TCP connection", "err", err)
				continue
			}
			go func() {
				defer conn.Close()
				c.processReader(conn)
			}()
		}
	}()

	udpAddress, err := net.ResolveUDPAddr("udp", *graphiteAddress)
	if err != nil {
		level.Error(logger).Log("msg", "Error resolving UDP address", "err", err)
		os.Exit(1)
	}
	udpSock, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		level.Error(logger).Log("msg", "Error listening to UDP address", "err", err)
		os.Exit(1)
	}
	go func() {
		defer udpSock.Close()
		for {
			buf := make([]byte, 65536)
			chars, srcAddress, err := udpSock.ReadFromUDP(buf)
			if err != nil {
				level.Error(logger).Log("msg", "Error reading UDP packet", "from", srcAddress, "err", err)
				continue
			}
			go c.processReader(bytes.NewReader(buf[0:chars]))
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`<html>
      <head><title>Graphite Exporter</title></head>
      <body>
      <h1>Graphite Exporter</h1>
      <p>Accepting plaintext Graphite samples over TCP and UDP on ` + *graphiteAddress + `</p>
      <p><a href="` + *metricsPath + `">Metrics</a></p>
      </body>
      </html>`))
	})

	level.Info(logger).Log("msg", "Listening on "+*listenAddress)
	level.Error(logger).Log("err", http.ListenAndServe(*listenAddress, nil))
	os.Exit(1)
}
