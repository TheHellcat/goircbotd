package main

import
(
    "fmt"
    "flag"
    "time"
    "strings"
    "os"
    "os/exec"
    "encoding/json"
    "hellcat/hcirc"
    "ircbotd/internal/ircbotint"
)

var cmdArgDebug bool
var cmdArgDaemon bool
var mainCtrl chan string
var shutdown bool = false
var running bool = true
var regedChatCommands []string
var regedTimedCommands []string

func init() {
    flag.BoolVar(&cmdArgDebug, "debug", false, "Enable debug-mode")
    flag.BoolVar(&cmdArgDaemon, "D", false, "Daemonize (launch into background)")
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

    sChatCommands = ""
    sTimedCommands = ""

    sJson = "{\"chatcommands\":[    {\"command\":\"!test1\"},    {\"command\":\"!test2\"},    {\"command\":\"!test3\"}],\"timedcommands\":[    {\"command\":\"-timertest1\", \"timer\":\"30\"},    {\"command\":\"-timertest2\", \"timer\":\"10\"},    {\"command\":\"-timertest3\", \"timer\":\"60\"}]}"

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

        // now lets split those temp. strings into nifty arrays holding all registered commands
        regedChatCommands = strings.Split(strings.Trim(sChatCommands, " "), " ")
        regedTimedCommands = strings.Split(strings.Trim(sTimedCommands, " "), " ")
    } else {
        // TODO: handle the error
    }
}


/**
 *
 */
func interfaceRegisteredCommand(command, channel, nick, user, host, text string) {

}


/**
 *
 */
func processPrivmsg(command, channel, nick, user, host, text string) {
    // TODO: - check if "command" is in the list of registered command
    // TODO: - call interfaceRegisteredCommand() in case it is
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
}


/**
 *
 */
func timedCommandsScheduler() {

}

func main() {

    var hcIrc *hcirc.HcIrc
    var cmd *exec.Cmd
    var err error
    var mainRunning bool
    var s string

    flag.Parse()

    // re-launch ourselfs as new process and quit if requested running as background daemon
    if cmdArgDaemon {
        fmt.Printf(": %s\n", os.Args[0])
        cmd = exec.Command(os.Args[0], "")
        err = cmd.Start()
        if err != nil {
            fmt.Printf("Error launching to background: %s\n", err.Error())
        } else {
            fmt.Printf("Successfully launched into background")
        }
        return
    }

    // fetch main config from parent application
    // TODO: fetch main config

    // fetch registered commands from parent application
    fetchRegisteredCommands()

    // some fancy "who am I splash" output :-)
    fmt.Printf("\n%s - %s\nfor %s\n%s\n\n", ircbotint.IrcBotName, ircbotint.IrcBotVersion,
        ircbotint.IrcBotParentProject, ircbotint.IrcBotC)

    // set up main control channel for communication from all worker-threads
    mainCtrl = make(chan string, 1)

    // flag to keep all worker threads running or tell them to exit
    running = true

    for !shutdown {

        hcIrc = hcirc.New("irc.hellcat.net", "6667", "Testuser", "Testnick", "")
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

            mainRunning = true
            for mainRunning {
                s = <-mainCtrl
                // TODO: handle messages from worker-threads
                s = s // silencing the compiler about currently unused variable
            }

        }

        hcIrc.Shutdown()
        hcIrc = nil

        if !shutdown {
            time.Sleep(time.Duration(10) * time.Second)
        }

    }

}
