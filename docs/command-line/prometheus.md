# prometheus

The Prometheus monitoring server

## Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help (also try --help-long and --help-man). |  |
| `--version` | Show application version. |  |
| `--config.file` | Prometheus configuration file path. | `prometheus.yml` |
| `--web.listen-address` | Address to listen on for UI, API, and telemetry. | `0.0.0.0:9090` |
| `--web.config.file` | [EXPERIMENTAL] Path to configuration file that can enable TLS or authentication. | `` |
| `--web.read-timeout` | Maximum duration before timing out read of the request, and closing idle connections. | `5m` |
| `--web.max-connections` | Maximum number of simultaneous connections. | `512` |
| `--web.external-url` | The URL under which Prometheus is externally reachable (for example, if Prometheus is served via a reverse proxy). Used for generating relative and absolute links back to Prometheus itself. If the URL has a path portion, it will be used to prefix all HTTP endpoints served by Prometheus. If omitted, relevant URL components will be derived automatically. |  |
| `--web.route-prefix` | Prefix for the internal routes of web endpoints. Defaults to path of --web.external-url. |  |
| `--web.user-assets` | Path to static asset directory, available at /user. |  |
| `--web.enable-lifecycle` | Enable shutdown and reload via HTTP request. | `false` |
| `--web.enable-admin-api` | Enable API endpoints for admin control actions. | `false` |
| `--web.enable-remote-write-receiver` | Enable API endpoint accepting remote write requests. | `false` |
| `--web.console.templates` | Path to the console template directory, available at /consoles. | `consoles` |
| `--web.console.libraries` | Path to the console library directory. | `console_libraries` |
| `--web.page-title` | Document title of Prometheus instance. | `Prometheus Time Series Collection and Processing Server` |
| `--web.cors.origin` | Regex for CORS origin. It is fully anchored. Example: 'https?://(domain1|domain2)\.com' | `.*` |
| `--storage.tsdb.path` | Base path for metrics storage. Use with server mode only. | `data/` |
| `--storage.tsdb.retention` | [DEPRECATED] How long to retain samples in storage. This flag has been deprecated, use "storage.tsdb.retention.time" instead. Use with server mode only. |  |
| `--storage.tsdb.retention.time` | How long to retain samples in storage. When this flag is set it overrides "storage.tsdb.retention". If neither this flag nor "storage.tsdb.retention" nor "storage.tsdb.retention.size" is set, the retention time defaults to 15d. Units Supported: y, w, d, h, m, s, ms. Use with server mode only. |  |
| `--storage.tsdb.retention.size` | Maximum number of bytes that can be stored for blocks. A unit is required, supported units: B, KB, MB, GB, TB, PB, EB. Ex: "512MB". Based on powers-of-2, so 1KB is 1024B. Use with server mode only. |  |
| `--storage.tsdb.no-lockfile` | Do not create lockfile in data directory. Use with server mode only. | `false` |
| `--storage.tsdb.head-chunks-write-queue-size` | Size of the queue through which head chunks are written to the disk to be m-mapped, 0 disables the queue completely. Experimental. Use with server mode only. | `0` |
| `--storage.agent.path` | Base path for metrics storage. Use with agent mode only. | `data-agent/` |
| `--storage.agent.wal-compression` | Compress the agent WAL. Use with agent mode only. | `true` |
| `--storage.agent.retention.min-time` | Minimum age samples may be before being considered for deletion when the WAL is truncated Use with agent mode only. |  |
| `--storage.agent.retention.max-time` | Maximum age samples may be before being forcibly deleted when the WAL is truncated Use with agent mode only. |  |
| `--storage.agent.no-lockfile` | Do not create lockfile in data directory. Use with agent mode only. | `false` |
| `--storage.remote.flush-deadline` | How long to wait flushing sample on shutdown or config reload. | `1m` |
| `--storage.remote.read-sample-limit` | Maximum overall number of samples to return via the remote read interface, in a single query. 0 means no limit. This limit is ignored for streamed response types. Use with server mode only. | `5e7` |
| `--storage.remote.read-concurrent-limit` | Maximum number of concurrent remote read calls. 0 means no limit. Use with server mode only. | `10` |
| `--storage.remote.read-max-bytes-in-frame` | Maximum number of bytes in a single frame for streaming remote read response types before marshalling. Note that client might have limit on frame size as well. 1MB as recommended by protobuf by default. Use with server mode only. | `1048576` |
| `--rules.alert.for-outage-tolerance` | Max time to tolerate prometheus outage for restoring "for" state of alert. Use with server mode only. | `1h` |
| `--rules.alert.for-grace-period` | Minimum duration between alert and restored "for" state. This is maintained only for alerts with configured "for" time greater than grace period. Use with server mode only. | `10m` |
| `--rules.alert.resend-delay` | Minimum amount of time to wait before resending an alert to Alertmanager. Use with server mode only. | `1m` |
| `--alertmanager.notification-queue-capacity` | The capacity of the queue for pending Alertmanager notifications. Use with server mode only. | `10000` |
| `--query.lookback-delta` | The maximum lookback duration for retrieving metrics during expression evaluations and federation. Use with server mode only. | `5m` |
| `--query.timeout` | Maximum time a query may take before being aborted. Use with server mode only. | `2m` |
| `--query.max-concurrency` | Maximum number of queries executed concurrently. Use with server mode only. | `20` |
| `--query.max-samples` | Maximum number of samples a single query can load into memory. Note that queries will fail if they try to load more samples than this into memory, so this also limits the number of samples a query can return. Use with server mode only. | `50000000` |
| `--enable-feature` | Comma separated feature names to enable. Valid options: agent, exemplar-storage, expand-external-labels, memory-snapshot-on-shutdown, promql-at-modifier, promql-negative-offset, promql-per-step-stats, remote-write-receiver (DEPRECATED), extra-scrape-metrics, new-service-discovery-manager, auto-gomaxprocs, no-default-scrape-port, native-histograms. See https://prometheus.io/docs/prometheus/latest/feature_flags/ for more details. | `` |
| `--log.level` | Only log messages with the given severity or above. One of: [debug, info, warn, error] | `info` |
| `--log.format` | Output format of log messages. One of: [logfmt, json] | `logfmt` |