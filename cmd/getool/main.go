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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/version"
)

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "Tooling for the Graphite Exporter.")
	app.Version(version.Print("getool"))
	app.HelpFlag.Short('h')

	defaultDBPath := "data/"

	importCmd := app.Command("create-blocks", "Import samples from OpenMetrics input and produce TSDB blocks. Please refer to the exporter docs for more details.")
	// TODO(aSquare14): add flag to set default block duration
	importFilePath := importCmd.Arg("whisper directory", "Directory of the whisper database.").Required().String()
	importDBPath := importCmd.Arg("output directory", "Output directory for generated blocks.").Default(defaultDBPath).String()
	importHumanReadable := importCmd.Flag("human-readable", "Print human readable values.").Short('r').Bool()
	importBlockDuration := importCmd.Flag("block-duration", "TSDB block duration.").Default("2h").Duration()
	importMappingConfig := importCmd.Flag("graphite.mapping-config", "Metric mapping configuration file name.").Default("").String()
	importStrictMatch := importCmd.Flag("graphite.mapping-strict-match", "Only import metrics that match the mapping configuration.").Bool()

	parsedCmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch parsedCmd {
	case importCmd.FullCommand():
		os.Exit(checkErr(backfillWhisper(*importFilePath, *importDBPath, *importMappingConfig, *importStrictMatch, *importHumanReadable, *importBlockDuration)))
	}
}

func checkErr(err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
