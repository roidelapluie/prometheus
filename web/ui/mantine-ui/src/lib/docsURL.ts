// Copyright The Prometheus Authors
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

// GLOBAL_VERSION is declared/defined in public/index.html, with its value
// replaced by Prometheus with the running version when serving the bundle.
declare const GLOBAL_VERSION: string;

// docsVersionPath returns the documentation path segment matching a Prometheus
// version string, e.g. "3.12" for "3.12.0". It falls back to "latest" when the
// version is empty or cannot be parsed (e.g. the unreplaced placeholder).
export const docsVersionPath = (version: string): string => {
  const match = /^(\d+\.\d+)/.exec(version || "");
  return match ? match[1] : "latest";
};

// prometheusDocsBaseURL is the base URL of the Prometheus documentation for the
// running version, with a trailing slash, e.g.
// "https://prometheus.io/docs/prometheus/3.12/".
export const prometheusDocsBaseURL = `https://prometheus.io/docs/prometheus/${docsVersionPath(
  typeof GLOBAL_VERSION === "undefined" ? "" : GLOBAL_VERSION
)}/`;
