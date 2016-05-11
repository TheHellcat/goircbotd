package hcirc

import (
    "fmt"
    "os"
    "bufio"
    "strings"
)


/**
 *
 */
func (hcIrc *HcIrc) StartConsole( ctrlChan chan string) {
    var ioReader *bufio.Reader
    var ioWriter *bufio.Writer
    var input string
    var inputLowered string
    var err error
    var consoleRunning bool

    ioReader = bufio.NewReader(os.Stdin)
    ioWriter = bufio.NewWriter(os.Stdout)

    fmt.Printf( "\nStarting interactive console:\n\n" )

// WIP NOT
fmt.Printf( "!!! Interactive console is currently pretty much useless!\n    It's only the base construct at the moment to add actual functionality, later!\n\n" )
// REMOVE ONCE THIS ACTUALLY DOES SOMETHING

    consoleRunning = true
    for consoleRunning {
        fmt.Print("> ")
        input, err = ioReader.ReadString('\n')
        input = strings.Trim( input, string('\n') )
        input = strings.Trim( input, " " )
        inputLowered = strings.ToLower( input )

        if nil == err {
            if "die" == inputLowered || "quit" == inputLowered || "exit" == inputLowered {
                ctrlChan <- "SHUTDOWN"
                consoleRunning = false
            } else if "" == input {
                // NOP
            } else if "restart" == inputLowered {
                ctrlChan <- "RESTART"
                consoleRunning = false
            } else if "help" == inputLowered {

            } else if "close" == inputLowered {
                consoleRunning = false
            } else {
                fmt.Printf( "Whooopsy, no idea what you're talking about there (i.e. command not recognized)\n" )
            }
        } else {
            consoleRunning = false
            fmt.Printf( "\n!!! closing console due to error: %s\n", err.Error() )
        }
    }

    fmt.Printf( "\nInteractive console terminated\n\n" )
}
