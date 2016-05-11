package main

import
(
    "fmt"
    "flag"
    "time"
    "strings"
    "strconv"
    "os"
    "os/exec"
    "encoding/json"
    "hellcat/hcirc"
    "ircbotd/internal/ircbotint"
)

var cmdArgDebug bool
var cmdArgDaemon bool
var cmdArgConsole bool
var mainCtrl chan string
var shutdown bool = false
var running bool = true
var regedChatCommands map[string]string
var regedTimedCommands map[string]int
var hcIrc *hcirc.HcIrc

func init() {
    flag.BoolVar(&cmdArgDebug, "debug", false, "Enable debug-mode")
    flag.BoolVar(&cmdArgDaemon, "D", false, "Daemonize (launch into background)")
    flag.BoolVar(&cmdArgConsole, "c", false, "Enable console (can not be used with -D)")
}


/**
 *
 */
func fetchRegisteredCommands() {
    var sJson string
    var jsonDecoder *json.Decoder
    var err error
    var jMap interface{}
    var jMapA map[string]interface{}
    var tmpMap map[string]interface{}
    var sChatCommands string
    var sTimedCommands string
    var a1 []string
    var a2 []string
    var i64 int64

    sChatCommands = ""
    sTimedCommands = ""

    sJson = "{\"chatcommands\":[    {\"command\":\"!test1\"},    {\"command\":\"!test2\"},    {\"command\":\"!test3\"},    {\"command\":\"!test4\"}],\"timedcommands\":[    {\"command\":\"-timertest1\", \"timer\":\"30\"},    {\"command\":\"-timertest2\", \"timer\":\"10\"},    {\"command\":\"-timertest3\", \"timer\":\"60\"}]}"

    jsonDecoder = json.NewDecoder(strings.NewReader(sJson))

    // first parse registered commands from the JSON response into temp. strings
    err = jsonDecoder.Decode(&jMap)
    if err == nil {
        jMapA = jMap.(map[string]interface{})
        for list, items := range jMapA {
            switch itemsT := items.(type) {
            case []interface{}:
                for _, itemVal := range itemsT {
                    tmpMap = itemVal.(map[string]interface{})
                    if "chatcommands" == list {
                        sChatCommands = fmt.Sprintf("%s %s", sChatCommands, tmpMap["command"])
                    }
                    if "timedcommands" == list {
                        sTimedCommands = fmt.Sprintf("%s %s*%s", sTimedCommands, tmpMap["command"], tmpMap["timer"])
                    }
                }
            }
        }

        // now lets write the commands into our maps
        a1 = strings.Split(strings.Trim(sChatCommands, " "), " ")
        for _, cmd := range a1 {
            regedChatCommands[cmd] = "1"
        }
        a1 = strings.Split(strings.Trim(sTimedCommands, " "), " ")
        for _, cmd := range a1 {
            a2 = strings.Split(cmd, "*")
            i64, _ = strconv.ParseInt(a2[1], 10, 32)
            regedTimedCommands[a2[0]] = int(i64)
        }
    } else {
        // TODO: handle the error
    }
}


/**
 *
 */
func interfaceRegisteredCommand(command, channel, nick, user, host, cmd, param string) {

    // test only
    if "!test1" == cmd {
        s := fmt.Sprintf("JOIN %s", param)
        hcIrc.OutQuickQueue <- s
    }
    if "!test2" == cmd {
        s := fmt.Sprintf("PRIVMSG %s :%s", channel, param)
        hcIrc.OutQuickQueue <- s
    }
    if "!test3" == cmd {
        mainCtrl <- "SHUTDOWN"
    }
    if "!test4" == cmd {
        mainCtrl <- "RESTART"
    }
    // test only

}


/**
 *
 */
func processPrivmsg(command, channel, nick, user, host, text string) {
    var isRegedChatCommand bool
    var a []string
    var cmd string
    var param string

    a = strings.SplitN(text, " ", 2)
    cmd = a[0]
    _, isRegedChatCommand = regedChatCommands[cmd]
    //fmt.Println()
    if len(a) == 2 {
        param = a[1]
    } else {
        param = ""
    }

    if isRegedChatCommand {
        go interfaceRegisteredCommand(command, channel, nick, user, host, cmd, param)
    }
}


/**
 * Main listener loop.
 * Processes and acts on messages received from the server
 */
func serverListener(hcIrc *hcirc.HcIrc) {
    var s string
    var command, channel, nick, user, host, text string

    for running {
        s = <-hcIrc.InboundQueue

        if len(hcIrc.Error) > 0 {
            // something bad happened - handle it!
            // TODO: handle the error
        } else {
            // all's good, process the message
            command, channel, nick, user, host, text = hcIrc.ParseMessage(s)

            if "PRIVMSG" == command {
                processPrivmsg(command, channel, nick, user, host, text)
            }
        }
    }

    if cmdArgDebug {
        fmt.Printf( "[LISTENERDEBUG] server listener thread ended" )
    }
}


/**
 *
 */
func timedCommandsScheduler() {

}

func main() {

    var cmd *exec.Cmd
    var err error
    var mainRunning bool
    var s string

    flag.Parse()

    // some fancy "who am I splash" output :-)
    fmt.Printf("\n%s - %s\n  for %s\n%s\n\n", ircbotint.IrcBotName, ircbotint.IrcBotVersion,
        ircbotint.IrcBotParentProject, ircbotint.IrcBotC)
    // TODO: make this super fancy :-D

    if cmdArgConsole && cmdArgDaemon {
        fmt.Printf( "ERROR: can not run as daemon/backgrounded with interactive console!\n       -c and -D can not be used simultaniously!\n\n" )
        return
    }

    // re-launch ourselfs as new process and quit if requested running as background daemon
    if cmdArgDaemon {
        fmt.Printf(": %s\n", os.Args[0])
        cmd = exec.Command(os.Args[0], "")
        err = cmd.Start()
        if err != nil {
            fmt.Printf("Error launching to background: %s\n\n", err.Error())
        } else {
            fmt.Printf("Successfully launched into background\n\n")
        }
        return
    }

    // set up main control channel for communication from all worker-threads
    mainCtrl = make(chan string, 1)

    for !shutdown {

        // flag to keep all worker threads running or tell them to exit
        running = true

        // fetch main config from parent application
        // TODO: fetch main config

        // init some stuff
        regedChatCommands = make(map[string]string)
        regedTimedCommands = make(map[string]int)

        // fetch registered commands from parent application
        fetchRegisteredCommands()

        hcIrc = hcirc.New("irc.hellcat.net", "6667", "Bottest", "Bottest", "")
        hcIrc.Debugmode = cmdArgDebug
        hcIrc.Connect()
        if len(hcIrc.Error) == 0 {

            // fire up server message queues
            hcIrc.StartInboundQueue()
            hcIrc.StartOutboundQueue()
            hcIrc.StartOutQuickQueue()

            // start main listener loop
            go serverListener(hcIrc)

            // start timed commands
            go timedCommandsScheduler()

            // start console if requested
            if cmdArgConsole {
                go hcIrc.StartConsole(mainCtrl)
            }

            mainRunning = true
            for mainRunning {
                s = <-mainCtrl

                if cmdArgDebug {
                    fmt.Printf( "[MAINDEBUG] received control command: %s\n", s )
                }

                if "SHUTDOWN" == s {
                    shutdown = true
                    running = false
                    mainRunning = false
                }
                if "RESTART" == s {
                    running = false
                    mainRunning = false
                }
            }

        }

        hcIrc.Shutdown()
        hcIrc = nil
        regedChatCommands = nil
        regedTimedCommands = nil

        if !shutdown {
            time.Sleep(time.Duration(10) * time.Second)
        }

    }

    fmt.Printf( "\nGood bye! Bot shutting down....\n\n" )
    time.Sleep(time.Duration(2) * time.Second)

}
