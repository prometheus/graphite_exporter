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
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIssue61(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	webAddr, graphiteAddr := fmt.Sprintf("127.0.0.1:%d", 9108), fmt.Sprintf("127.0.0.1:%d", 9109)
	exporter := exec.Command(
		filepath.Join(cwd, "..", "graphite_exporter"),
		"--web.listen-address", webAddr,
		"--graphite.listen-address", graphiteAddr,
		"--graphite.mapping-config", filepath.Join(cwd, "fixtures", "mapping.yml"),
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

	now := time.Now()

	input := `rspamd.actions.add_header 2 NOW
rspamd.actions.greylist 0 NOW
rspamd.actions.no_action 24 NOW
rspamd.actions.reject 1 NOW
rspamd.actions.rewrite_subject 0 NOW
rspamd.actions.soft_reject 0 NOW
rspamd.bytes_allocated 4165268944 NOW
rspamd.chunks_allocated 4294966730 NOW
rspamd.chunks_freed 0 NOW
rspamd.chunks_oversized 1 NOW
rspamd.connections 1 NOW
rspamd.control_connections 1 NOW
rspamd.ham_count 24 NOW
rspamd.learned 2 NOW
rspamd.pools_allocated 59 NOW
rspamd.pools_freed 171 NOW
rspamd.scanned 27 NOW
rspamd.shared_chunks_allocated 34 NOW
rspamd.spam_count 3 NOW`
	input = strings.NewReplacer("NOW", fmt.Sprintf("%d", now.Unix())).Replace(input)

	conn, err := net.Dial("tcp", graphiteAddr)
	if err != nil {
		t.Fatalf("connection error: %v", err)
	}
	defer conn.Close()
	_, err = conn.Write([]byte(input))
	if err != nil {
		t.Fatalf("write error: %v", err)
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
	for _, s := range []string{"rspamd_actions{action=\"add_header\"} 2", "rspamd_connections 1"} {
		if !strings.Contains(string(b), s) {
			t.Fatalf("Expected %q in %q – input: %q – time: %s", s, string(b), input, now)
		}
	}
}
