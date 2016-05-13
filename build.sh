#!/bin/bash

rm -Rf bin
rm -Rf pkg
# export GOROOT=/data1/build/go
export GOPATH=$PWD
export GOBIN=$PWD/bin
go clean src/ircbotd/ircbotd.go
go install src/ircbotd/ircbotd.go
