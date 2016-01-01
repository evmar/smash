#!/bin/sh

set -x

# This script installs gocairo by running its makefile.
# This is different than "go install ..." because the makefile
# generates the binding code based on the symbols available in
# the locally-installed cairo headers.

REPO=github.com/martine/gocairo

mkdir -p $GOPATH/src
cd $GOPATH/src
git clone --depth=1 git://$REPO $REPO
cd $REPO

# install c2go.
# This doesn't work:
#   go get -u
# Fails with:
#   package _/home/travis/gopath/github.com/martine/gocairo: unrecognized import path "_/home/travis/gopath/github.com/martine/gocairo"
go get "rsc.io/c2go/cc"

rm cairo/cairo.go  # to force rerunning the generate script
make

