package hcirc

import (
    "fmt"
    "bufio"
    "strings"
    "io"
)

type consoleCommandCallback func(string, string) string

var consoleRegisteredCommands map[string]consoleCommandCallback
var consoleRegisteredCommandInfos map[string]string


/**
 *
 */
func (hcIrc *HcIrc) RegisterConsoleCommand(command string, description string, function consoleCommandCallback) {
    command = strings.ToLower(command)
    consoleRegisteredCommands[command] = function
    consoleRegisteredCommandInfos[command] = description
}


/**
 *
 */
func (hcIrc *HcIrc) StartConsole(ctrlChan chan string, ioRead io.Reader, ioWrite io.Writer) {
    var ioReader *bufio.Reader
    var ioWriter *bufio.Writer
    var input string
    var inputLowered string
    var err error
    var consoleRunning bool
    var output string
    var exists bool
    var function consoleCommandCallback
    var a []string
    var s string
    var t string

    ioReader = bufio.NewReader(ioRead)
    ioWriter = bufio.NewWriter(ioWrite)

    fmt.Printf("\nStarting interactive console:\n\n")

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
            if "" == input {
                // NOP
            } else if "die" == inputLowered {
                ctrlChan <- "SHUTDOWN"
                consoleRunning = false
            } else if "restart" == inputLowered {
                ctrlChan <- "RESTART"
                consoleRunning = false
            } else if "help" == inputLowered {
                s = fmt.Sprintf("Internal commands:\n  die  - shuts down the bot\n  retart - restarts the bot\n  exit - quits the console\n\n")
                s = fmt.Sprintf("%sRegistered commands:\n", s)
                for t, _ = range consoleRegisteredCommands {
                    s = fmt.Sprintf("%s  %s - %s\n", s, t, consoleRegisteredCommandInfos[t])
                }
                ioWriter.WriteString(s)
            } else if "exit" == inputLowered {
                consoleRunning = false
            } else {
                a = strings.SplitN(inputLowered, " ", 2)
                function, exists = consoleRegisteredCommands[a[0]]
                if exists {
                    a = strings.SplitN(input, " ", 2)
                    if len(a) > 1 {
                        s = a[1]
                    } else {
                        s = ""
                    }
                    s = function(a[0], s)
                    ioWriter.WriteString(s)
                } else {
                    output = fmt.Sprintf("Whooopsy, no idea what you're talking about there (i.e. command not recognized)\nMaybe try \"help\".\n")
                    ioWriter.WriteString(output)
                }
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
