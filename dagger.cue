package main

import (
	"dagger.io/dagger"
	"dagger.io/dagger/core"
	"github.com/roidelapluie/parazonium/parazonium"
)

dagger.#Plan & {
    client: filesystem: ".": read: {
        contents: dagger.#FS
    }
	actions: {
		test:  parazonium.#Build & {"client": client, cmd: "make test GO_ONLY=1"}
		ui:  parazonium.#Build & {"client": client, cmd: "make assets-tarball ui-lint ui-test"}
		build: parazonium.#Build & {"client": client, cmd: "make build", binaries: ["prometheus"]}
		all:   core.#Nop & {input: [test.output, build.output, ui.output]}
	}
}
