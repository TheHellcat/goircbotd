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
    "hellcat/hcthreadutils"
    "ircbotd/internal/ircbotint"
    "ircbotd/internal/ircbotext"
)

type strMainConfig struct {
    botNick     string
    botUsername string
    botRealname string

    netHost     string
    netPort     string
    netPassword string
    netChannels []string
}

var cmdArgDebug bool
var cmdArgDaemon bool
var cmdArgConsole bool
var cmdArgUrl string
var cmdArgStandalone string
var cmdArgWebsocketBind string
var cmdArgWebsocketEnabled bool
var cmdArgDatadir string
var mainCtrl chan string
var shutdown bool = false
var running bool = true
var regedChatCommands map[string]string
var regedTimedCommands map[string]int
var hcIrc *hcirc.HcIrc
var listenerThreadId string
var timedcommandsThreadId string
var runningStandalone bool

var mainConfig strMainConfig

func init() {
    flag.BoolVar(&cmdArgDebug, "debug", false, "Enable debug-mode")
    flag.BoolVar(&cmdArgDaemon, "D", false, "Daemonize (launch into background)")
    flag.BoolVar(&cmdArgConsole, "c", false, "Enable console (can not be used with -D)")
    flag.StringVar(&cmdArgUrl, "base", "", "Base URL for accessing parent application")
    flag.StringVar(&cmdArgStandalone, "standalone", "", "Enable stand-alone mode and load main config from given file")
    flag.BoolVar(&cmdArgWebsocketEnabled, "ws", false, "Enable listening for incomming http/websocket connections")
    flag.StringVar(&cmdArgWebsocketBind, "wsbind", "0.0.0.0:8088", "Listen binding for incomming http/websocket connections (defaults to 0.0.0.0:8088)")
    flag.StringVar(&cmdArgDatadir, "datadir", ".", "Directory to save datafiles. Defaults to current dir.")
}


/**
 *
 */
func fetchMainConfig() (bool, string) {
    var ok bool
    var notok string
    var err error
    var rJson string
    var jMap interface{}
    var jMapA map[string]interface{}
    var jsonDecoder *json.Decoder

    ok = false
    notok = "unknown reason (this really should not happen, but it just did)"

    runningStandalone = false
    if len(cmdArgStandalone) > 0 {
        runningStandalone = true
    }

    if runningStandalone {
        // TODO: load config from file
    } else {
        if len(cmdArgUrl) > 10 {
            ircbotint.SetHttpUrl(cmdArgUrl)
            rJson, err = ircbotint.CallHttp("getmainconfig", "")
            if err == nil {
                ok = true
            } else {
                notok = fmt.Sprintf("Failed to fetch config: %s", err.Error())
            }
        } else {
            notok = "No or incomplete base URL specified\n    Use --base=<URL> or --standalone=<PATH> to supply config location"
        }
    }

    if !ok {
        fmt.Printf("(!) ERROR loading config: %s\n", notok)
    }

    // parse JSON response into config vars
    jsonDecoder = json.NewDecoder(strings.NewReader(rJson))
    err = jsonDecoder.Decode(&jMap)
    if err == nil {
        jMapA = jMap.(map[string]interface{})
        for k, v := range jMapA {
            switch itemT := v.(type) {
            case string:
                if ( "botNick" == k ) {
                    mainConfig.botNick = itemT
                } else if ( "botUsername" == k ) {
                    mainConfig.botUsername = itemT
                } else if ( "botRealname" == k ) {
                    mainConfig.botRealname = itemT
                } else if ( "netHost" == k ) {
                    mainConfig.netHost = itemT
                } else if ( "netPort" == k ) {
                    mainConfig.netPort = itemT
                } else if ( "netChannels" == k ) {
                    mainConfig.netChannels = strings.Split(itemT, " ")
                } else if ( "netPassword" == k ) {
                    mainConfig.netPassword = itemT
                }
            }
        }
    }

    return ok, notok
}


/**
 *
 */
func chatCmdFuncCB(command, channel, nick, user, host, cmd, param string) string {
    go interfaceRegisteredCommand(command, channel, nick, user, host, cmd, param)

    // interfaceRegisteredCommand() takes care of sending output to server queues by itself
    return ""
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

    sJson, err = ircbotint.CallHttp("getchatcommands", "")
    if err != nil {
        fmt.Printf("(!) ERROR fetching chat commands: %s\n", err.Error())
        return
    }

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
            ircbotint.RegisterInternalChatCommand(cmd, chatCmdFuncCB)
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
        hcIrc.OutboundQueue <- s
    }
    if "!test3" == cmd {
        mainCtrl <- "SHUTDOWN"
    }
    if "!test4" == cmd {
        mainCtrl <- "RESTART"
    }
    if "!test5" == cmd {
        for joined := range hcIrc.JoinedChannels {
            s := fmt.Sprintf("PRIVMSG %s :I am in %s", channel, joined)
            hcIrc.OutboundQueue <- s
        }
    }
    // test only

}


/**
 *
 */
func processPrivmsg(command, channel, nick, user, host, text string) {
    //var isRegedChatCommand bool
    var a []string
    var cmd string
    var param string

    a = strings.SplitN(text, " ", 2)
    cmd = a[0]
    if len(a) == 2 {
        param = a[1]
    } else {
        param = ""
    }

    /* not needed anymore, call to interfaceRegisteredCommand() is now handled by the registered callback */
    //_, isRegedChatCommand = regedChatCommands[cmd]
    //if isRegedChatCommand {
    //    go interfaceRegisteredCommand(command, channel, nick, user, host, cmd, param)
    //}

    // fun fact: which (the registered commands handler) is initiated by this very call xD
    ircbotint.HandleCommand(command, channel, nick, user, host, cmd, param)
}


/**
 * Main listener loop.
 * Processes and acts on messages received from the server
 */
func serverListener(hcIrc *hcirc.HcIrc) {
    var s string
    var command, channel, nick, user, host, text string

    listenerThreadId = hcthreadutils.GetRoutineId()
    if cmdArgDebug {
        fmt.Printf("[LISTENERDEBUG] server listener thread started, ID=%s\n", listenerThreadId)
    }

    hcIrc.AutohandleSysMsgs = false

    for running {
        s = <-hcIrc.InboundQueue

        if len(hcIrc.Error) > 0 {
            // something bad happened - handle it!
            // TODO: handle the error
        } else {
            // all's good, process the message
            command, channel, nick, user, host, text = hcIrc.ParseMessage(s)
            hcIrc.HandleSystemMessages(command, channel, nick, user, host, text, s)

            if "PRIVMSG" == command {
                processPrivmsg(command, channel, nick, user, host, text)
            }
        }
    }

    if cmdArgDebug {
        fmt.Printf("[LISTENERDEBUG] server listener thread ended\n")
    }
}


/**
 *
 */
func timedCommandsScheduler() {
    timedcommandsThreadId = hcthreadutils.GetRoutineId()
    if cmdArgDebug {
        fmt.Printf("[TIMERDEBUG] command scheduler thread started, ID=%s\n", timedcommandsThreadId)
    }

    if cmdArgDebug {
        fmt.Printf("[TIMERDEBUG] command scheduler thread ended\n")
    }
}

func main() {

    var cmd *exec.Cmd
    var err error
    var mainRunning bool
    var s string
    var b bool

    flag.Parse()

    // some fancy "who am I splash" output :-)
    fmt.Printf("\n%s - %s\n  for %s\n%s\n\n", ircbotint.IrcBotName, ircbotint.IrcBotVersion,
        ircbotint.IrcBotParentProject, ircbotint.IrcBotC)
    // TODO: make this super fancy :-D

    if cmdArgConsole && cmdArgDaemon {
        fmt.Printf("ERROR: can not run as daemon/backgrounded with interactive console!\n       -c and -D can not be used simultaniously!\n\n")
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
        // regedConsoleCommands = make(map[string]string)

        // fetch main configuration from parent application
        b, s = fetchMainConfig()
        if !b {
            return
        }

        hcIrc = hcirc.New(mainConfig.netHost, mainConfig.netPort, mainConfig.botUsername, mainConfig.botNick, mainConfig.netPassword)
        hcIrc.SetRealname(mainConfig.botRealname)
        hcIrc.Debugmode = cmdArgDebug
        hcIrc.SetDataDir( cmdArgDatadir )
        hcIrc.Connect()
        if len(hcIrc.Error) == 0 {

            // register webserver / websockets listener console commands and start the listener, if requested
            ircbotint.RegisterWebsocketConsoleCommands()
            if cmdArgWebsocketEnabled {
                go ircbotint.EnableWebsocketServer(hcIrc, cmdArgWebsocketBind)
            }

            fmt.Printf("(i) Connected to %s:%s\n", mainConfig.netHost, mainConfig.netPort)

            // register console commands
            hcIrc.RegisterAdditionalConsoleCommands()

            // fire up server message queues
            hcIrc.StartInboundQueue()
            hcIrc.StartOutboundQueue()
            hcIrc.StartOutQuickQueue()

            // init handler for internal chat commands
            ircbotint.InitChatcmdHan(hcIrc)

            // fetch registered commands from parent application
            fetchRegisteredCommands()

            // start main listener loop
            go serverListener(hcIrc)

            // start timed commands
            go timedCommandsScheduler()

            // init all configured extensions
            ircbotext.InitExtensions(hcIrc)

            // join all configured auto-join channels
            for _, s = range mainConfig.netChannels {
                s = fmt.Sprintf("JOIN %s", s)
                hcIrc.OutQuickQueue <- s
            }

            // start console if requested
            if cmdArgConsole {
                go hcIrc.StartConsole(mainCtrl, os.Stdin, os.Stdout)
            }

            mainRunning = true
            for mainRunning {
                s = <-mainCtrl

                if cmdArgDebug {
                    fmt.Printf("[MAINDEBUG] received control command: %s\n", s)
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

        } else {
            fmt.Printf("(!) Failed to connecto to %s:%s - %s\n", mainConfig.netHost, mainConfig.netPort, hcIrc.Error)
        }

        // give all active exentions the chance to clean up
        ircbotext.ShutdownExtensions(hcIrc)

        // shut down http/ws listener
        ircbotint.DisableWebsocketServer()

        hcIrc.Shutdown()
        hcthreadutils.WaitForRoutinesEndById([]string{listenerThreadId, timedcommandsThreadId})
        hcIrc = nil
        regedChatCommands = nil
        regedTimedCommands = nil

        if !shutdown {
            if cmdArgDebug {
                fmt.Printf("[MAINDEBUG] waiting 10 seconds for restart/reconnect\n")
            }
            time.Sleep(time.Duration(10) * time.Second)
        }

    }

    fmt.Printf("\nGood bye! Bot shutting down....\n\n")
    time.Sleep(time.Duration(2) * time.Second)

}
