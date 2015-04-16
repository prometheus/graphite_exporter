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
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	listeningAddress = flag.String("web.listen-address", ":9108", "Address on which to expose metrics.")
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "Path under which to expose Prometheus metrics.")
	graphiteAddress  = flag.String("graphite.listen-address", ":9109", "TCP and UDP address on which to accept samples.")
	mappingConfig    = flag.String("graphite.mapping-config", "", "Metric mapping configuration file name.")
	sampleExpiry     = flag.Duration("graphite.sample-expiry", 5*time.Minute, "How long a sample is valid for.")
	lastProcessed    = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "graphite_last_processed_timestamp_seconds",
			Help: "Unix timestamp of the last processed graphite metric.",
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

type graphiteCollector struct {
	samples map[string]*graphiteSample
	mu      *sync.Mutex
	mapper  *metricMapper
	ch      chan *graphiteSample
}

func newGraphiteCollector() *graphiteCollector {
	c := &graphiteCollector{
		ch:      make(chan *graphiteSample, 0),
		mu:      &sync.Mutex{},
		samples: map[string]*graphiteSample{},
	}
	go c.processSamples()
	return c
}

func (c *graphiteCollector) processReader(reader io.Reader) {
	lineScanner := bufio.NewScanner(reader)
	for {
		if ok := lineScanner.Scan(); !ok {
			break
		}
		c.processLine(lineScanner.Text())
	}
}

func (c *graphiteCollector) processLine(line string) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		glog.Infof("Invalid part count of %d in line: %s", len(parts), line)
		return
	}
	var name string
	labels, present := c.mapper.getMapping(parts[0])
	if present {
		name = labels["name"]
		delete(labels, "name")
	} else {
		name = invalidMetricChars.ReplaceAllString(parts[0], "_")
	}

	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		glog.Infof("Invalid value in line: %s", line)
		return
	}
	timestamp, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		glog.Infof("Invalid timestamp in line: %s", line)
		return
	}
	sample := graphiteSample{
		OriginalName: parts[0],
		Name:         name,
		Value:        value,
		Labels:       labels,
		Type:         prometheus.GaugeValue,
		Help:         fmt.Sprintf("Graphite metric %s", name),
		Timestamp:    time.Unix(int64(timestamp), int64(math.Mod(timestamp, 1.0)*1e9)),
	}
	lastProcessed.Set(float64(time.Now().UnixNano()) / 1e9)
	c.ch <- &sample
}

func (c *graphiteCollector) processSamples() {
	ticker := time.NewTicker(time.Minute).C
	for {
		select {
		case sample := <-c.ch:
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

func main() {
	flag.Parse()
	http.Handle(*metricsPath, prometheus.Handler())
	c := newGraphiteCollector()
	prometheus.MustRegister(c)

	c.mapper = &metricMapper{}
	if *mappingConfig != "" {
		err := c.mapper.initFromFile(*mappingConfig)
		if err != nil {
			glog.Fatalf("Error loading metric mapping config: %s", err)
		}
	}

	tcpSock, err := net.Listen("tcp", *graphiteAddress)
	if err != nil {
		glog.Fatalf("Error binding to TCP socket: %s", err)
	}
	go func() {
		for {
			conn, err := tcpSock.Accept()
			if err != nil {
				glog.Errorf("Error accepting TCP connection: %s", err)
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
		glog.Fatalf("Error resolving UDP address: %s", err)
	}
	udpSock, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		glog.Fatalf("Error listening to UDP address: %s", err)
	}
	go func() {
		defer udpSock.Close()
		for {
			buf := make([]byte, 65536)
			chars, srcAddress, err := udpSock.ReadFromUDP(buf)
			if err != nil {
				glog.Errorf("Error reading UDP packet from %s: %s", srcAddress, err)
				continue
			}
			go c.processReader(bytes.NewReader(buf[0:chars]))
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
      <head><title>Graphite Exporter</title></head>
      <body>
      <h1>Graphite Exporter</h1>
      <p>Accepting plaintext Graphite samples over TCP and UDP on ` + *graphiteAddress + `</p>
      <p><a href="` + *metricsPath + `">Metrics</a></p>
      </body>
      </html>`))
	})
	http.ListenAndServe(*listeningAddress, nil)
}
