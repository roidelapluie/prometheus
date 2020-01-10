// Copyright 2020 The Prometheus Authors
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/prometheus/tsdb/testutil"
)

// queryLogLine is a basic representation of a query log line.
type queryLogLine map[string]interface{}

func enableQueryLog(t *testing.T, configFile *os.File, queryLogFile string) {
	err := configFile.Truncate(0)
	testutil.Ok(t, err)
	_, err = configFile.Seek(0, 0)
	testutil.Ok(t, err)
	_, err = configFile.Write([]byte(fmt.Sprintf("global:\n  query_log_file: %s\n", queryLogFile)))
	testutil.Ok(t, err)
	postReloadConfig(t)
}

func postReloadConfig(t *testing.T) {
	r, err := http.Post("http://127.0.0.1:9090/-/reload", "text/plain", nil)
	testutil.Ok(t, err)
	testutil.Equals(t, 200, r.StatusCode)
}

func disableQueryLog(t *testing.T, configFile *os.File) {
	err := configFile.Truncate(0)
	testutil.Ok(t, err)
	postReloadConfig(t)
}

func readQueryLog(t *testing.T, path string) []queryLogLine {
	ql := []queryLogLine{}
	file, err := os.Open(path)
	testutil.Ok(t, err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var q queryLogLine
		testutil.Ok(t, json.Unmarshal(scanner.Bytes(), &q))
		ql = append(ql, q)
	}
	return ql
}
func TestQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	queryLogFile, err := ioutil.TempFile("", "query")
	testutil.Ok(t, err)
	defer os.Remove(queryLogFile.Name())
	configFile, err := ioutil.TempFile("", "config")
	testutil.Ok(t, err)
	defer os.Remove(queryLogFile.Name())

	prom := exec.Command(promPath, "--config.file="+configFile.Name(), "--web.enable-lifecycle")
	testutil.Ok(t, prom.Start())
	defer func() {
		prom.Process.Signal(os.Interrupt)
		prom.Wait()
	}()

	testutil.Ok(t, waitForPrometheus())

	enableQueryLog(t, configFile, queryLogFile.Name())
	testutil.Equals(t, 0, len(readQueryLog(t, queryLogFile.Name())))

	_, err = http.Get("http://127.0.0.1:9090/api/v1/query?query=time()")
	testutil.Ok(t, err)
	ql := readQueryLog(t, queryLogFile.Name())
	testutil.Equals(t, 1, len(ql))
	testutil.Equals(t, ql[0]["query"].(string), "time()")
	testutil.Equals(t, ql[0]["clientIP"].(string), "127.0.0.1")
	testutil.Equals(t, ql[0]["path"].(string), "/api/v1/query")
	testutil.Equals(t, ql[0]["method"].(string), "GET")

	_, err = http.Get("http://[::1]:9090/api/v1/query?query=vector(1)%20")
	testutil.Ok(t, err)
	ql = readQueryLog(t, queryLogFile.Name())
	testutil.Equals(t, 2, len(ql))
	testutil.Equals(t, ql[1]["query"].(string), "vector(1) ")
	testutil.Equals(t, ql[1]["clientIP"].(string), "::1")

	disableQueryLog(t, configFile)
	_, err = http.Get("http://127.0.0.1:9090/api/v1/query?query=time()")
	testutil.Ok(t, err)
	ql = readQueryLog(t, queryLogFile.Name())
	testutil.Equals(t, 2, len(ql))

	enableQueryLog(t, configFile, queryLogFile.Name())
	_, err = http.Get("http://127.0.0.1:9090/api/v1/query?query=time()")
	testutil.Ok(t, err)
	ql = readQueryLog(t, queryLogFile.Name())
	testutil.Equals(t, 3, len(ql))

	// Move the file, Prometheus should still write to the old file.
	newFile, err := ioutil.TempFile("", "newLoc")
	testutil.Ok(t, err)
	defer os.Remove(newFile.Name())
	testutil.Ok(t, os.Rename(queryLogFile.Name(), newFile.Name()))
	ql = readQueryLog(t, newFile.Name())
	testutil.Equals(t, 3, len(ql))

	_, err = http.Get("http://127.0.0.1:9090/api/v1/query?query=time()")
	testutil.Ok(t, err)
	ql = readQueryLog(t, newFile.Name())
	testutil.Equals(t, 4, len(ql))

	// Reload config, Prometheus should write to the new file..
	postReloadConfig(t)
	_, err = http.Get("http://127.0.0.1:9090/api/v1/query?query=time()")
	testutil.Ok(t, err)
	ql = readQueryLog(t, newFile.Name())
	testutil.Equals(t, 4, len(ql))
	ql = readQueryLog(t, queryLogFile.Name())
	testutil.Equals(t, 1, len(ql))
}

func waitForPrometheus() error {
	var err error
	for x := 0; x < 10; x++ {
		// error=nil means prometheus has started so can send the interrupt signal and wait for the grace shutdown.
		if _, err = http.Get("http://localhost:9090/graph"); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}
