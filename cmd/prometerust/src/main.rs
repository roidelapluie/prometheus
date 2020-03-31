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

extern crate clap;
use clap::{Arg, App};

fn main() {
    let _matches = App::new("prometheus")
        .about("The prometheus monitoring system")
        .arg(Arg::with_name("config.file")
             .long("config.file")
             .takes_value(true)
             .help("Prometheus configuration file path.")
             .default_value("prometheus.yml"))
        .arg(Arg::with_name("web.listen-address")
             .long("web.listen-address")
             .takes_value(true)
             .help("Address to listen on for UI, API, and telemetry.")
             .default_value("0.0.0.0:9090"))
        .arg(Arg::with_name("web.read-timeout")
             .long("web.read-timeout")
             .takes_value(true)
             .help("Maximum duration before timing out read of the request, and closing idle connections.")
             .default_value("5m"))
        .arg(Arg::with_name("web.max-connections")
             .long("web.max-connections")
             .takes_value(true)
             .help("Maximum number of simultaneous connections.")
             .default_value("512"))
        .get_matches();
}
