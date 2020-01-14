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
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/prometheus/util/testutil"
)

type origin int

const (
	apiOrigin origin = iota
	consoleOrigin
	ruleOrigin
)

type queryLogTestParams struct {
	origin         origin
	prefix         string
	host           string
	port           int
	enabledAtStart bool
}

type queryLogLine map[string]interface{}

func queryLogTest(t *testing.T, p *queryLogTestParams) {
	// setup temporary files for this test
	queryLogFile, err := ioutil.TempFile("", "query")
	testutil.Ok(t, err)
	defer os.Remove(queryLogFile.Name())
	configFile, err := ioutil.TempFile("", "config")
	testutil.Ok(t, err)
	defer os.Remove(configFile.Name())

	if p.enabledAtStart {
		enableQueryLog(t, p, configFile, queryLogFile.Name(), false)
	} else {
		disableQueryLog(t, p, configFile, false)
	}

	params := append([]string{"--config.file=" + configFile.Name(), "--web.enable-lifecycle", fmt.Sprintf("--web.listen-address=%s:%d", p.host, p.port)}, p.params()...)
	prom := exec.Command(promPath, params...)
	testutil.Ok(t, prom.Start())
	defer func() {
		prom.Process.Signal(os.Interrupt)
		prom.Wait()
	}()
	testutil.Ok(t, waitForPrometheus(p))

	if !p.enabledAtStart {
		runQuery(t, p)
		testutil.Equals(t, 0, len(readQueryLog(t, queryLogFile.Name())))
		enableQueryLog(t, p, configFile, queryLogFile.Name(), true)
	}

	runQuery(t, p)

	ql := readQueryLog(t, queryLogFile.Name())
	qc := len(ql)
	if p.exactQueryCount() {
		testutil.Equals(t, 1, qc)
	} else {
		testutil.Assert(t, qc > 0, "no queries logged")
	}
	validateLastQuery(t, p, ql)

	disableQueryLog(t, p, configFile, true)
	if !p.exactQueryCount() {
		qc = len(readQueryLog(t, queryLogFile.Name()))
	}

	runQuery(t, p)

	ql = readQueryLog(t, queryLogFile.Name())
	testutil.Equals(t, qc, len(ql))

	qc = len(ql)
	enableQueryLog(t, p, configFile, queryLogFile.Name(), true)

	runQuery(t, p)
	qc++

	ql = readQueryLog(t, queryLogFile.Name())
	if p.exactQueryCount() {
		testutil.Equals(t, qc, len(ql))
	} else {
		testutil.Assert(t, len(ql) > qc, "no queries logged")
	}
	validateLastQuery(t, p, ql)
	qc = len(ql)

	// Move the file, Prometheus should still write to the old file.
	newFile, err := ioutil.TempFile("", "newLoc")
	testutil.Ok(t, err)
	defer os.Remove(newFile.Name())
	testutil.Ok(t, os.Rename(queryLogFile.Name(), newFile.Name()))
	ql = readQueryLog(t, newFile.Name())
	if p.exactQueryCount() {
		testutil.Equals(t, qc, len(ql))
	}
	validateLastQuery(t, p, ql)
	qc = len(ql)

	runQuery(t, p)

	qc++

	ql = readQueryLog(t, newFile.Name())
	if p.exactQueryCount() {
		testutil.Equals(t, qc, len(ql))
	} else {
		testutil.Assert(t, len(ql) > qc, "no queries logged")
	}
	validateLastQuery(t, p, ql)

	postReloadConfig(t, p)

	runQuery(t, p)

	ql = readQueryLog(t, queryLogFile.Name())
	qc = len(ql)
	if p.exactQueryCount() {
		testutil.Equals(t, 1, qc)
	} else {
		testutil.Assert(t, qc > 0, "no queries logged")
	}

}

func waitForPrometheus(p *queryLogTestParams) error {
	var err error
	for x := 0; x < 10; x++ {
		// error=nil means prometheus has started so can send
		// the interrupt signal and wait for the grace shutdown.
		if _, err = http.Get(fmt.Sprintf("http://%s:%d%s/graph", p.host, p.port, p.prefix)); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

func enableQueryLog(t *testing.T, p *queryLogTestParams, configFile *os.File, queryLogFile string, reload bool) {
	err := configFile.Truncate(0)
	testutil.Ok(t, err)
	_, err = configFile.Seek(0, 0)
	testutil.Ok(t, err)
	_, err = configFile.Write([]byte(fmt.Sprintf("global:\n  query_log_file: %s\n", queryLogFile)))
	testutil.Ok(t, err)
	_, err = configFile.Write([]byte(p.configuration()))
	testutil.Ok(t, err)
	if reload {
		postReloadConfig(t, p)
	}
}

func disableQueryLog(t *testing.T, p *queryLogTestParams, configFile *os.File, reload bool) {
	err := configFile.Truncate(0)
	testutil.Ok(t, err)
	_, err = configFile.Seek(0, 0)
	testutil.Ok(t, err)
	_, err = configFile.Write([]byte(p.configuration()))
	testutil.Ok(t, err)
	if reload {
		postReloadConfig(t, p)
	}
}

func postReloadConfig(t *testing.T, p *queryLogTestParams) {
	r, err := http.Post(fmt.Sprintf("http://%s:%d%s/-/reload", p.host, p.port, p.prefix), "text/plain", nil)
	testutil.Ok(t, err)
	testutil.Equals(t, 200, r.StatusCode)
}

func runQuery(t *testing.T, p *queryLogTestParams) {
	switch p.origin {
	case apiOrigin:
		apiQuery(t, p)
	case consoleOrigin:
		consoleQuery(t, p)
	case ruleOrigin:
		ruleQuery(t, p)
	default:
		panic("can't query this origin")
	}
}

func skip(t *testing.T, p *queryLogTestParams) (bool, string) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", p.host))
	if err != nil {
		return true, "ip version not supported"
	}
	l.Close()
	return false, ""
}

func apiQuery(t *testing.T, p *queryLogTestParams) {
	_, err := http.Get(fmt.Sprintf(
		"http://%s:%d%s/api/v1/query?query=%s",
		p.host,
		p.port,
		p.prefix,
		url.QueryEscape("query_with_api"),
	))
	testutil.Ok(t, err)
}

func consoleQuery(t *testing.T, p *queryLogTestParams) {
	_, err := http.Get(fmt.Sprintf(
		"http://%s:%d%s/consoles/test.html",
		p.host,
		p.port,
		p.prefix,
	))
	testutil.Ok(t, err)
}

func ruleQuery(t *testing.T, p *queryLogTestParams) {
	time.Sleep(2 * time.Second)
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

func (p *queryLogTestParams) String() string {
	var name string
	switch p.origin {
	case apiOrigin:
		name = "api queries"
	case consoleOrigin:
		name = "console queries"
	case ruleOrigin:
		name = "rule queries"
	}
	name = name + ", " + p.host
	if p.enabledAtStart {
		name = name + ", enabled at start"
	}
	if p.prefix != "" {
		name = name + ", with prefix " + p.prefix
	}
	return name
}

func (p *queryLogTestParams) params() []string {
	s := []string{}
	if p.prefix != "" {
		s = append(s, "--web.route-prefix="+p.prefix)
	}
	if p.origin == consoleOrigin {
		s = append(s, "--web.console.templates="+filepath.Join("testdata", "consoles"))
	}
	return s
}

func validateLastQuery(t *testing.T, p *queryLogTestParams, ql []queryLogLine) {
	q := ql[len(ql)-1]
	testutil.Equals(t, q["query"].(string), p.query())
	switch p.origin {
	case consoleOrigin:
		testutil.Equals(t, q["path"].(string), p.prefix+"/consoles/test.html")
	case apiOrigin:
		testutil.Equals(t, q["path"].(string), p.prefix+"/api/v1/query")
	case ruleOrigin:
		testutil.Equals(t, q["groupName"].(string), "querylogtest")
		testutil.Equals(t, q["groupFile"].(string), filepath.Join(cwd, "testdata", "rules", "test.yml"))
	default:
		panic("unknown origin")
	}
	if p.origin != ruleOrigin {
		host := p.host
		if host == "[::1]" {
			host = "::1"
		}
		testutil.Equals(t, q["clientIP"].(string), host)
	}
}

func (p *queryLogTestParams) query() string {
	switch p.origin {
	case apiOrigin:
		return "query_with_api"
	case ruleOrigin:
		return "query_in_rule"
	case consoleOrigin:
		return "query_in_console"
	default:
		panic("unknown origin")
	}
}

func (p *queryLogTestParams) configuration() string {
	switch p.origin {
	case ruleOrigin:
		return "\nrule_files:\n- " + filepath.Join(cwd, "testdata", "rules", "test.yml") + "\n"
	default:
		return "\n"
	}
}

func (p *queryLogTestParams) exactQueryCount() bool {
	return p.origin != ruleOrigin
}

func TestQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	port := 15000
	for _, origin := range []origin{apiOrigin, consoleOrigin, ruleOrigin} {
		for _, host := range []string{"127.0.0.1", "[::1]"} {
			for _, enabledAtStart := range []bool{true, false} {
				for _, prefix := range []string{"", "/foobar"} {
					p := &queryLogTestParams{
						origin:         origin,
						host:           host,
						enabledAtStart: enabledAtStart,
						prefix:         prefix,
						port:           port,
					}

					t.Run(p.String(), func(t *testing.T) {
						if skip, msg := skip(t, p); skip {
							t.Skip(msg)
						}
						queryLogTest(t, p)
					})
				}
			}
		}
	}
}
