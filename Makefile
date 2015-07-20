.PHONY: all vendor test godoc cover

all:
	go install -race smash

vendor:
	go install -tags gtk_3_10 github.com/conformal/gotk3/gtk

test:
	go test smash

godoc:
	godoc -http=:6060 &

cover:
	go test -coverprofile=cov
	go tool cover -html=cov
