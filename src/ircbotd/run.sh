#!/bin/bash

OLDPWD=$PWD
cd ../..
# export GOROOT=
export GOPATH=$PWD
cd $OLDPWD
go clean ircbotd.go
go run ircbotd.go --debug
