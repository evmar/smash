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
	protoc \
	  -Iproto \
	  --plugin="protoc-gen-ts=$(protoc_gen_ts)" \
	  --go_out="cli/proto" \
	  --js_out="import_style=commonjs,binary:web/js" \
	  --ts_out="web/src" \
	  proto/smash.proto

.PHONY: fmt
fmt:
	./fmt.sh --write

.PHONY: serve
serve:
	cd web && node js/server.js
