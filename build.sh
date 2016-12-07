#!/bin/bash

rm -Rf bin
rm -Rf pkg
export GOPATH=$PWD
export GOBIN=$PWD/bin

V1=$(git log -n 1 --format=format:"%ci;%ct;%H;%h" HEAD)
V2=$(git branch | grep '*' | cut -d ' ' -f2)
V="$V1;$V2"
sed -i "s#{_GITVER_}#$V#g" src/ircbotd/internal/ircbotint/version.go

go clean github.com/mattn/go-sqlite3
go clean src/ircbotd/ircbotd.go
go install github.com/mattn/go-sqlite3
go install src/ircbotd/ircbotd.go
