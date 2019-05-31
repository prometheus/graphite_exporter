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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
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
	invalidMetricChars = regexp.MustCompile("[^a-zA-Z0-9_:]")
)

type graphiteSample struct {
	OriginalName string
	Name         string
	Labels       map[string]string
	Help         string
	Value        float64
	Type         prometheus.ValueType
	Timestamp    time.Time
}

type metricMapper interface {
	GetMapping(string, mapper.MetricType) (*mapper.MetricMapping, prometheus.Labels, bool)
	InitFromFile(string) error
}

type graphiteCollector struct {
	samples     map[string]*graphiteSample
	mu          *sync.Mutex
	mapper      metricMapper
	sampleCh    chan *graphiteSample
	lineCh      chan string
	strictMatch bool
}

func newGraphiteCollector() *graphiteCollector {
	c := &graphiteCollector{
		sampleCh:    make(chan *graphiteSample),
		lineCh:      make(chan string),
		mu:          &sync.Mutex{},
		samples:     map[string]*graphiteSample{},
		strictMatch: *strictMatch,
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

func (c *graphiteCollector) processLine(line string) {
	line = strings.TrimSpace(line)
	log.Debugf("Incoming line : %s", line)
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		log.Infof("Invalid part count of %d in line: %s", len(parts), line)
		return
	}
	originalName := parts[0]
	var name string
	mapping, labels, present := c.mapper.GetMapping(originalName, mapper.MetricTypeGauge)

	if (present && mapping.Action == mapper.ActionTypeDrop) || (!present && c.strictMatch) {
		return
	}

	if present {
		name = invalidMetricChars.ReplaceAllString(mapping.Name, "_")
	} else {
		name = invalidMetricChars.ReplaceAllString(originalName, "_")
	}

	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		log.Infof("Invalid value in line: %s", line)
		return
	}
	timestamp, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		log.Infof("Invalid timestamp in line: %s", line)
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
	log.Debugf("Sample: %+v", sample)
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

func dumpFSM(mapper *mapper.MetricMapper, dumpFilename string) error {
	f, err := os.Create(dumpFilename)
	if err != nil {
		return err
	}
	log.Infoln("Start dumping FSM to", dumpFilename)
	w := bufio.NewWriter(f)
	mapper.FSM.DumpFSM(w)
	w.Flush()
	f.Close()
	log.Infoln("Finish dumping FSM")
	return nil
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("graphite_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	prometheus.MustRegister(sampleExpiryMetric)
	sampleExpiryMetric.Set(sampleExpiry.Seconds())

	log.Infoln("Starting graphite_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	c := newGraphiteCollector()
	prometheus.MustRegister(c)

	c.mapper = &mapper.MetricMapper{}
	if *mappingConfig != "" {
		err := c.mapper.InitFromFile(*mappingConfig)
		if err != nil {
			log.Fatalf("Error loading metric mapping config: %s", err)
		}
	}

	if *dumpFSMPath != "" {
		err := dumpFSM(c.mapper.(*mapper.MetricMapper), *dumpFSMPath)
		if err != nil {
			log.Fatal("Error dumping FSM:", err)
		}
	}

	tcpSock, err := net.Listen("tcp", *graphiteAddress)
	if err != nil {
		log.Fatalf("Error binding to TCP socket: %s", err)
	}
	go func() {
		for {
			conn, err := tcpSock.Accept()
			if err != nil {
				log.Errorf("Error accepting TCP connection: %s", err)
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
		log.Fatalf("Error resolving UDP address: %s", err)
	}
	udpSock, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Fatalf("Error listening to UDP address: %s", err)
	}
	go func() {
		defer udpSock.Close()
		for {
			buf := make([]byte, 65536)
			chars, srcAddress, err := udpSock.ReadFromUDP(buf)
			if err != nil {
				log.Errorf("Error reading UDP packet from %s: %s", srcAddress, err)
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

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
