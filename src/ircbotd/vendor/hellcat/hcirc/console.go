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
func (hcIrc *HcIrc) StartConsole(ctrlChan chan string) {
    var ioReader *bufio.Reader
    var ioWriter *bufio.Writer
    var input string
    var inputLowered string
    var err error
    var consoleRunning bool
    var output string

    ioReader = bufio.NewReader(os.Stdin)
    ioWriter = bufio.NewWriter(os.Stdout)

    fmt.Printf("\nStarting interactive console:\n\n")

    // WIP NOTE
    fmt.Printf("!!! Interactive console is currently pretty much useless!\n    It's only the base construct at the moment to add actual functionality, later!\n\n")
    // REMOVE ONCE THIS ACTUALLY DOES SOMETHING

    consoleRunning = true
    for consoleRunning {
        output = fmt.Sprintf("> ")
        ioWriter.WriteString(output)
        ioWriter.Flush()
        input, err = ioReader.ReadString('\n')
        input = strings.Trim(input, string('\n'))
        input = strings.Trim(input, " ")
        inputLowered = strings.ToLower(input)

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
                output = fmt.Sprintf("Whooopsy, no idea what you're talking about there (i.e. command not recognized)\n")
                ioWriter.WriteString(output)
            }
        } else {
            consoleRunning = false
            output = fmt.Sprintf("\n!!! closing console due to error: %s\n", err.Error())
            ioWriter.WriteString(output)
        }
        ioWriter.Flush()
    }

    output = fmt.Sprintf("\nInteractive console terminated\n\n")
    ioWriter.WriteString(output)
    ioWriter.Flush()
}
