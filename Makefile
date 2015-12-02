.PHONY: all test godoc cov

all:
	go install -race smash

test:
	go test smash/...

godoc:
	godoc -http=:6060 &

cov:
	go test -coverprofile=cov
	go tool cover -html=cov
