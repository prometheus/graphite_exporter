// Copyright 2018 The Prometheus Authors
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

package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestIssue90(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	webAddr, graphiteAddr := fmt.Sprintf("127.0.0.1:%d", 9118), fmt.Sprintf("127.0.0.1:%d", 9119)
	exporter := exec.Command(
		filepath.Join(cwd, "..", "graphite_exporter"),
		"--graphite.mapping-strict-match",
		"--web.listen-address", webAddr,
		"--graphite.listen-address", graphiteAddr,
		"--graphite.mapping-config", filepath.Join(cwd, "fixtures", "issue90.yml"),
	)
	err = exporter.Start()
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	defer exporter.Process.Kill()

	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		resp, err := http.Get("http://" + webAddr)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			break
		}
	}

	testInputs, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures", "issue90_in.txt"))
	if err != nil {
		t.Fatalf("failed to read input fixture: %v", err)
	}

	lines := bytes.Split(testInputs, []byte{'\n'})
	currSec := time.Now().Unix() - 2
	for _, input := range lines {
		conn, err := net.Dial("udp", graphiteAddr)
		updateInput := bytes.ReplaceAll(input, []byte("NOW"), []byte(strconv.FormatInt(currSec, 10)))

		if err != nil {
			t.Fatalf("connection error: %v", err)
		}
		_, err = conn.Write(updateInput)
		if err != nil {
			t.Fatalf("write error: %v", err)
		}
		conn.Close()
	}

	resp, err := http.Get("http://" + path.Join(webAddr, "metrics"))
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("unexpected status, want 200, got %v, body: %s", resp.Status, b)
	}

}
