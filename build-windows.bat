@echo off

set GOPATH=%CD%
set GOBIN=%CD%\bin

go clean src\github.com\josephspurrier\goversioninfo\cmd\goversioninfo\goversioninfo.go
go clean github.com\mattn\go-sqlite3
go install src\github.com\josephspurrier\goversioninfo\cmd\goversioninfo\goversioninfo.go
go install github.com\mattn\go-sqlite3

go clean src/ircbotd/ircbotd.go

go generate src/ircbotd/ircbotd.go
move src\ircbotd\resource.syso src\ircbotd\internal\ircbotint\resource.syso
go install src/ircbotd/ircbotd.go

del src\ircbotd\internal\ircbotint\resource.syso
