package parazonium

import (
	"encoding/yaml"
	"strconv"
	"dagger.io/dagger/core"
	"universe.dagger.io/docker"
)

#Build: {
	client: _
	cmd:    string
	_promu: core.#ReadFile & {
		input: client.filesystem.".".read.contents
		path:  ".promu.yml"
	}
	_goVersion: strconv.FormatFloat(yaml.Unmarshal(_promu.contents).go.version, 102, 2, 64)
	_builder:   docker.#Pull & {
		source: "quay.io/prometheus/golang-builder:" + _goVersion + "-base"
	}
	docker.#Run & {
		input: _builder.output
		entrypoint: ["/bin/bash", "-c"]
		command: name: cmd
		env: {
			GOMODCACHE:       _modCachePath
			npm_config_cache: _npmCachePath
		}
		_modCachePath:   "/go-mod-cache"
		_buildCachePath: "/go-build-cache"
		_npmCachePath:   "/npm-build-cache"
		mounts: {
			app: {
				dest:     "/app"
				contents: client.filesystem.".".read.contents
			}
			"go mod cache": {
				contents: core.#CacheDir & {
					id: "go_mod"
				}
				dest: _modCachePath
			}
			"npm cache": {
				contents: core.#CacheDir & {
					id: "npm"
				}
				dest: _npmCachePath
			}
			"go build cache": {
				contents: core.#CacheDir & {
					id: "go_build"
				}
				dest: _buildCachePath
			}
		}
		workdir: "/app"
	}
}
