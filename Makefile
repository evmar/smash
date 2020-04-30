protoc_gen_ts ?= ./web/node_modules/.bin/protoc-gen-ts

.PHONY: run
run: cli/smash
	cd cli && ./smash

cli/smash: cli/proto/smash.go
	cd cli && go build github.com/evmar/smash/cmd/smash

.PHONY: web
web: tsc
	cd web && yarn run webpack $(arg)

.PHONY: tsc
tsc: web/src/proto.ts
	cd web && yarn run tsc $(arg)

# Build the proto generator from the TypeScript source.
proto/gen.js: proto/*.ts
	cd proto && yarn run tsc $(arg)

# Build the proto output using the proto generator.
web/src/proto.ts: proto/gen.js proto/smash.d.ts
	node proto/gen.js ts proto/smash.d.ts > web/src/proto.ts
cli/proto/smash.go: proto/gen.js proto/smash.d.ts
	node proto/gen.js go proto/smash.d.ts > cli/proto/smash.go

# Target to manually run proto generation.
.PHONY: proto
proto: web/src/proto.ts cli/proto/smash.go

.PHONY: fmt
fmt:
	./fmt.sh --write
	go fmt ./cli/...

.PHONY: serve
serve:
	cd web && node js/server.js
