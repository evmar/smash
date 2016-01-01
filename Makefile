.PHONY: all test godoc cov

all:
	go install -race .

test:
	go test ./...

godoc:
	godoc -http=:6060 &

cov:
	go test -coverprofile=cov
	go tool cover -html=cov
