#!/bin/sh

set -x

# This script installs gocairo by running its makefile.
# This is different than "go install ..." because the makefile
# generates the binding code based on the symbols available in
# the locally-installed cairo headers.

REPO=github.com/martine/gocairo

cd $GOPATH
git clone --depth=1 git://$REPO $REPO
cd $REPO
echo $GOPATH
pwd
go get -u
rm cairo/cairo.go  # to force rerunning the generate script
make

