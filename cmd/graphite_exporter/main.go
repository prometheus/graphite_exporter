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
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	clientVersion "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
	"github.com/prometheus/statsd_exporter/pkg/mappercache/lru"
	"github.com/prometheus/statsd_exporter/pkg/mappercache/randomreplacement"

	"github.com/prometheus/graphite_exporter/collector"
)

var (
	metricsPath     = kingpin.Flag("web.telemetry-path", "Path under which to expose Prometheus metrics.").Default("/metrics").String()
	graphiteAddress = kingpin.Flag("graphite.listen-address", "TCP and UDP address on which to accept samples.").Default(":9109").String()
	mappingConfig   = kingpin.Flag("graphite.mapping-config", "Metric mapping configuration file name.").Default("").String()
	sampleExpiry    = kingpin.Flag("graphite.sample-expiry", "How long a sample is valid for.").Default("5m").Duration()
	strictMatch     = kingpin.Flag("graphite.mapping-strict-match", "Only store metrics that match the mapping configuration.").Bool()
	cacheSize       = kingpin.Flag("graphite.cache-size", "Maximum size of your metric mapping cache. Relies on least recently used replacement policy if max size is reached.").Default("1000").Int()
	cacheType       = kingpin.Flag("graphite.cache-type", "Metric mapping cache type. Valid options are \"lru\" and \"random\"").Default("lru").Enum("lru", "random")
	dumpFSMPath     = kingpin.Flag("debug.dump-fsm", "The path to dump internal FSM generated for glob matching as Dot file.").Default("").String()
	checkConfig     = kingpin.Flag("check-config", "Check configuration and exit.").Default("false").Bool()
	toolkitFlags    = kingpinflag.AddFlags(kingpin.CommandLine, ":9108")
)

func init() {
	prometheus.MustRegister(clientVersion.NewCollector("graphite_exporter"))
}

func dumpFSM(mapper *mapper.MetricMapper, dumpFilename string, logger *slog.Logger) error {
	if mapper.FSM == nil {
		return fmt.Errorf("no FSM available to be dumped, possibly because the mapping contains regex patterns")
	}
	f, err := os.Create(dumpFilename)
	if err != nil {
		return err
	}
	logger.Info("Start dumping FSM", "to", dumpFilename)
	w := bufio.NewWriter(f)
	mapper.FSM.DumpFSM(w)
	w.Flush()
	f.Close()
	logger.Info("Finish dumping FSM")
	return nil
}

func main() {
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("graphite_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promslog.New(promslogConfig)

	logger.Info("Starting graphite_exporter", "version_info", version.Info())
	logger.Info(version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	c := collector.NewGraphiteCollector(logger, *strictMatch, *sampleExpiry)
	prometheus.MustRegister(c)

	metricMapper := &mapper.MetricMapper{Logger: logger}
	if *mappingConfig != "" {
		err := metricMapper.InitFromFile(*mappingConfig)
		if err != nil {
			logger.Error("Error loading metric mapping config", "err", err)
			os.Exit(1)
		}
	}

	cache, err := getCache(*cacheSize, *cacheType, prometheus.DefaultRegisterer)
	if err != nil {
		logger.Error("error initializing mapper cache", "err", err)
		os.Exit(1)
	}
	metricMapper.UseCache(cache)

	if *checkConfig {
		logger.Info("Configuration check successful, exiting")
		return
	}

	if *dumpFSMPath != "" {
		err := dumpFSM(metricMapper, *dumpFSMPath, logger)
		if err != nil {
			logger.Error("Error dumping FSM", "err", err)
			os.Exit(1)
		}
	}

	c.SetMapper(metricMapper)

	tcpSock, err := net.Listen("tcp", *graphiteAddress)
	if err != nil {
		logger.Error("Error binding to TCP socket", "err", err)
		os.Exit(1)
	}
	go func() {
		for {
			conn, err := tcpSock.Accept()
			if err != nil {
				logger.Error("Error accepting TCP connection", "err", err)
				continue
			}
			go func() {
				defer conn.Close()
				c.ProcessReader(conn)
			}()
		}
	}()

	udpAddress, err := net.ResolveUDPAddr("udp", *graphiteAddress)
	if err != nil {
		logger.Error("Error resolving UDP address", "err", err)
		os.Exit(1)
	}
	udpSock, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		logger.Error("Error listening to UDP address", "err", err)
		os.Exit(1)
	}
	go func() {
		defer udpSock.Close()
		for {
			buf := make([]byte, 65536)
			chars, srcAddress, err := udpSock.ReadFromUDP(buf)
			if err != nil {
				logger.Error("Error reading UDP packet", "from", srcAddress, "err", err)
				continue
			}
			go c.ProcessReader(bytes.NewReader(buf[0:chars]))
		}
	}()

	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			Name:        "Graphite Exporter",
			Description: "Prometheus Graphite Exporter",
			ExtraHTML:   `<p>Accepting plaintext Graphite samples over TCP and UDP on ` + *graphiteAddress + `</p>`,
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error("error running HTTP server", "err", err)
		os.Exit(1)
	}
}

// TODO(mr): this is copied verbatim from statsd_exporter/main.go. It should be a
// convenience function in mappercache, but that caused an import cycle.
func getCache(cacheSize int, cacheType string, registerer prometheus.Registerer) (mapper.MetricMapperCache, error) {
	var cache mapper.MetricMapperCache
	var err error
	if cacheSize == 0 {
		return nil, nil
	} else {
		switch cacheType {
		case "lru":
			cache, err = lru.NewMetricMapperLRUCache(registerer, cacheSize)
		case "random":
			cache, err = randomreplacement.NewMetricMapperRRCache(registerer, cacheSize)
		default:
			err = fmt.Errorf("unsupported cache type %q", cacheType)
		}

		if err != nil {
			return nil, err
		}
	}

	return cache, nil
}
