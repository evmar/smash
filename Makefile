.PHONY: all run
all: cli/smash web/dist/smash.js

run: all
	cd cli && ./smash

cli/smash: cli/cmd/smash/*.go cli/proto/smash.go cli/vt100/terminal.go cli/bash/aliases.go cli/bash/complete.go
	cd cli && go build github.com/evmar/smash/cmd/smash

webts=$(wildcard web/src/*.ts)

web/dist/smash.js: web/package.json $(webts)
	web/node_modules/.bin/esbuild --target=es2019 --bundle --sourcemap --outfile=$@ web/src/smash.ts

# Build the proto generator from the TypeScript source.
proto/gen.js: proto/*.ts
	cd proto && yarn run tsc

# Build the proto output using the proto generator.
web/src/proto.ts: proto/gen.js proto/smash.d.ts
	node proto/gen.js ts proto/smash.d.ts > web/src/proto.ts
cli/proto/smash.go: proto/gen.js proto/smash.d.ts
	node proto/gen.js go proto/smash.d.ts > cli/proto/smash.go

.PHONY: watch
watch:
	(cd proto && yarn run tsc --preserveWatchOutput -w & \
	cd web && yarn run tsc --preserveWatchOutput -w & \
	wait)

.PHONY: fmt
fmt:
	./fmt.sh --write
	go fmt ./cli/...

.PHONY: serve
serve:
	cd web && node js/server.js
