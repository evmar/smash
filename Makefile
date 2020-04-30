protoc_gen_ts ?= ./web/node_modules/.bin/protoc-gen-ts

.PHONY: cli
cli:
	cd cli && go run github.com/evmar/smash/cmd/smash

.PHONY: web
web:
	cd web && yarn run webpack $(arg)

.PHONY: tsc
tsc:
	cd web && yarn run tsc $(arg)

.PHONY: proto
proto:
	node proto/gen.js ts proto/smash.d.ts > web/src/proto.ts
	node proto/gen.js go proto/smash.d.ts > cli/proto/smash.go

.PHONY: fmt
fmt:
	./fmt.sh --write

.PHONY: serve
serve:
	cd web && node js/server.js
