.PHONY: all vendor test godoc cov basic

all:
	go install -race smash

vendor:
	go install -tags gtk_3_10 github.com/conformal/gotk3/gtk

test:
	go test smash

godoc:
	godoc -http=:6060 &

cov:
	go test -coverprofile=cov
	go tool cover -html=cov

basic:
	go build shell/basic/basic.go
