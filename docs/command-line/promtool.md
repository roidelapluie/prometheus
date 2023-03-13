---
title: promtool

---

# promtool

Tooling for the Prometheus monitoring system.

## Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help (also try --help-long and --help-man). |  |
| `--version` | Show application version. |  |
| `--enable-feature` | Comma separated feature names to enable (only PromQL related and no-default-scrape-port). See https://prometheus.io/docs/prometheus/latest/feature_flags/ for the options and more details. | `` |## Commands

| Command | Description |
| --- | --- |
| help | Show help. |
| check | Check the resources for validity. |
| query | Run query against a Prometheus server. |
| debug | Fetch debug information. |
| test | Unit testing. |
| tsdb | Run tsdb commands. |

### `promtool help`

Show help.

### Arguments

| Argument | Description |
| --- | --- |
| command | Show help on command. |

### `promtool check`

Check the resources for validity.

### Flags

| Flag | Description |
| --- | --- |
| `--extended` | Print extended information related to the cardinality of the metrics. |

### `promtool check service-discovery`

Perform service discovery for the given job name and report the results, including relabeling.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--timeout` | The time to wait for discovery results. | `30s` |### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| config-file | The prometheus config file. | Yes |
| job | The job to run service discovery for. | Yes |

### `promtool check config`

Check if the config files are valid or not.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--syntax-only` | Only check the config file syntax, ignoring file and content validation referenced in the config |  |
| `--lint` | Linting checks to apply to the rules specified in the config. Available options are: all, duplicate-rules, none. Use --lint=none to disable linting | `duplicate-rules` |
| `--lint-fatal` | Make lint errors exit with exit code 3. | `false` |
| `--agent` | Check config file for Prometheus in Agent mode. |  |### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| config-files | The config files to check. | Yes |

### `promtool check web-config`

Check if the web config files are valid or not.

### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| web-config-files | The config files to check. | Yes |

### `promtool check rules`

Check if the rule files are valid or not.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--lint` | Linting checks to apply. Available options are: all, duplicate-rules, none. Use --lint=none to disable linting | `duplicate-rules` |
| `--lint-fatal` | Make lint errors exit with exit code 3. | `false` |### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| rule-files | The rule files to check. | Yes |

### `promtool check metrics`

Pass Prometheus metrics over stdin to lint them for consistency and correctness.

examples:

$ cat metrics.prom | promtool check metrics

$ curl -s http://localhost:9090/metrics | promtool check metrics



### `promtool query`

Run query against a Prometheus server.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-o`, `--format` | Output format of the query. | `promql` |
| `--http.config.file` | HTTP client configuration file for promtool to connect to Prometheus. |  |

### `promtool query instant`

Run instant query.

### Flags

| Flag | Description |
| --- | --- |
| `--time` | Query evaluation time (RFC3339 or Unix timestamp). |### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| server | Prometheus server to query. | Yes |
| expr | PromQL query expression. | Yes |

### `promtool query range`

Run range query.

### Flags

| Flag | Description |
| --- | --- |
| `--header` | Extra headers to send to server. |
| `--start` | Query range start time (RFC3339 or Unix timestamp). |
| `--end` | Query range end time (RFC3339 or Unix timestamp). |
| `--step` | Query step size (duration). |### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| server | Prometheus server to query. | Yes |
| expr | PromQL query expression. | Yes |

### `promtool query series`

Run series query.

### Flags

| Flag | Description |
| --- | --- |
| `--match` | Series selector. Can be specified multiple times. |
| `--start` | Start time (RFC3339 or Unix timestamp). |
| `--end` | End time (RFC3339 or Unix timestamp). |### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| server | Prometheus server to query. | Yes |

### `promtool query labels`

Run labels query.

### Flags

| Flag | Description |
| --- | --- |
| `--start` | Start time (RFC3339 or Unix timestamp). |
| `--end` | End time (RFC3339 or Unix timestamp). |
| `--match` | Series selector. Can be specified multiple times. |### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| server | Prometheus server to query. | Yes |
| name | Label name to provide label values for. | Yes |

### `promtool debug`

Fetch debug information.



### `promtool debug pprof`

Fetch profiling debug information.

### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| server | Prometheus server to get pprof files from. | Yes |

### `promtool debug metrics`

Fetch metrics debug information.

### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| server | Prometheus server to get metrics from. | Yes |

### `promtool debug all`

Fetch all debug information.

### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| server | Prometheus server to get all debug information from. | Yes |

### `promtool test`

Unit testing.



### `promtool test rules`

Unit tests for rules.

### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| test-rule-file | The unit test file. | Yes |

### `promtool tsdb`

Run tsdb commands.



### `promtool tsdb bench`

Run benchmarks.



### `promtool tsdb bench write`

Run a write performance benchmark.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--out` | Set the output path. | `benchout` |
| `--metrics` | Number of metrics to read. | `10000` |
| `--scrapes` | Number of scrapes to simulate. | `3000` |### Arguments

| Argument | Description | Default |
| --- | --- | --- |
| file | Input file with samples data, default is (../../tsdb/testdata/20kseries.json). | `../../tsdb/testdata/20kseries.json` |

### `promtool tsdb analyze`

Analyze churn, label pair cardinality and compaction efficiency.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--limit` | How many items to show in each list. | `20` |
| `--extended` | Run extended analysis. |  |### Arguments

| Argument | Description | Default |
| --- | --- | --- |
| db path | Database path (default is data/). | `data/` |
| block id | Block to analyze (default is the last block). |  |

### `promtool tsdb list`

List tsdb blocks.

### Flags

| Flag | Description |
| --- | --- |
| `-r`, `--human-readable` | Print human readable values. |### Arguments

| Argument | Description | Default |
| --- | --- | --- |
| db path | Database path (default is data/). | `data/` |

### `promtool tsdb dump`

Dump samples from a TSDB.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--min-time` | Minimum timestamp to dump. | `-9223372036854775808` |
| `--max-time` | Maximum timestamp to dump. | `9223372036854775807` |
| `--match` | Series selector. | `{__name__=~'(?s:.*)'}` |### Arguments

| Argument | Description | Default |
| --- | --- | --- |
| db path | Database path (default is data/). | `data/` |

### `promtool tsdb create-blocks-from`

[Experimental] Import samples from input and produce TSDB blocks. Please refer to the storage docs for more details.

### Flags

| Flag | Description |
| --- | --- |
| `-r`, `--human-readable` | Print human readable values. |
| `-q`, `--quiet` | Do not print created blocks. |

### `promtool tsdb create-blocks-from openmetrics`

Import samples from OpenMetrics input and produce TSDB blocks. Please refer to the storage docs for more details.

### Arguments

| Argument | Description | Default | Required |
| --- | --- | --- | --- |
| input file | OpenMetrics file to read samples from. |  | Yes |
| output directory | Output directory for generated blocks. | `data/` |  |

### `promtool tsdb create-blocks-from rules`

Create blocks of data for new recording rules.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--http.config.file` | HTTP client configuration file for promtool to connect to Prometheus. |  |
| `--url` | The URL for the Prometheus API with the data where the rule will be backfilled from. | `http://localhost:9090` |
| `--start` | The time to start backfilling the new rule from. Must be a RFC3339 formatted date or Unix timestamp. Required. |  |
| `--end` | If an end time is provided, all recording rules in the rule files provided will be backfilled to the end time. Default will backfill up to 3 hours ago. Must be a RFC3339 formatted date or Unix timestamp. |  |
| `--output-dir` | Output directory for generated blocks. | `data/` |
| `--eval-interval` | How frequently to evaluate rules when backfilling if a value is not set in the recording rule files. | `60s` |### Arguments

| Argument | Description | Required |
| --- | --- | --- |
| rule-files | A list of one or more files containing recording rules to be backfilled. All recording rules listed in the files will be backfilled. Alerting rules are not evaluated. | Yes |