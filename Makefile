flatc ?= ~/projects/flatbuffers/flatc

.PHONY: cli
cli:
	cd cli && go run github.com/evmar/smash/cmd/smash

.PHONY: web
web:
	cd web && yarn run webpack -w

.PHONY: proto
proto:
	$(flatc) --go -o cli proto/smash.fbs
	$(flatc) --ts --short-names --no-fb-import -o web/src proto/smash.fbs
