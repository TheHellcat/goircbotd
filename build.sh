#!/bin/bash

rm -Rf bin
rm -Rf pkg
export GOPATH=$PWD
export GOBIN=$PWD/bin

go clean github.com/mattn/go-sqlite3
go clean src/ircbotd/ircbotd.go
go install github.com/mattn/go-sqlite3
go install src/ircbotd/ircbotd.go
