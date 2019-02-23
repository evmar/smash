protoc_gen_ts ?= ./web/node_modules/.bin/protoc-gen-ts

.PHONY: cli
cli:
	cd cli && go run github.com/evmar/smash/cmd/smash

.PHONY: web
web:
	cd web && yarn run webpack -w

.PHONY: proto
proto:
	protoc \
	  -Iproto \
	  --plugin="protoc-gen-ts=$(protoc_gen_ts)" \
	  --go_out="cli/proto" \
	  --js_out="import_style=commonjs,binary:web/src" \
	  --ts_out="web/src" \
	  proto/smash.proto
