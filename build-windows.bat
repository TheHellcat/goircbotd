@echo off

set GOPATH=%CD%
set GOBIN=%CD%\bin

go install src\github.com\josephspurrier\goversioninfo\cmd\goversioninfo\goversioninfo.go
go generate src/ircbotd/ircbotd.go
move src\ircbotd\resource.syso src\ircbotd\internal\ircbotint\resource.syso
go install src/ircbotd/ircbotd.go

del src\ircbotd\internal\ircbotint\resource.syso
