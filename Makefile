.PHONY: run
run: cli/smash
	cd cli && ./smash

cli/smash: cli/proto/smash.go web/dist/smash.bundle.js
	cd cli && go build github.com/evmar/smash/cmd/smash

web/dist/smash.bundle.js: web/src/*.ts
	cd web && yarn run tsc
	cd web && yarn run webpack -p

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
	cd web && yarn run webpack -w & \
	wait)

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
