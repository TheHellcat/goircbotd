package main

import
(
    "fmt"
    "flag"
    "time"
    "os"
    "os/exec"
    "hellcat/hcirc"
    "ircbotd/internal/ircbotint"
)

var cmdArgDebug bool
var cmdArgDaemon bool
var mainCtrl chan string
var shutdown bool = false
var running bool = true


func init() {
    flag.BoolVar(&cmdArgDebug, "debug", false, "Enable debug-mode")
    flag.BoolVar(&cmdArgDaemon, "D", false, "Daemonize (launch into background)")
}


/**
 * Main listener loop.
 * Processes and acts on messages received from the server
 */
func serverListener(hcIrc *hcirc.HcIrc) {
    var s string
    var command, channel, nick, user, host, text string

    for running {
        s = <- hcIrc.InboundQueue

        if len(hcIrc.Error) > 0 {
            // something bad happened - handle it!
            // TODO: handle the error
        } else {
            // all's good, process the message
            command, channel, nick, user, host, text = hcIrc.ParseMessage(s)
            fmt.Printf( "%s, %s, %s, %s, %s, %s\n", command, channel, nick, user, host, text )
        }
    }
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
        fmt.Printf( ": %s\n", os.Args[0] )
        cmd = exec.Command( os.Args[0], "" )
        err = cmd.Start()
        if err != nil {
            fmt.Printf( "Error launching to background: %s\n", err.Error() )
        } else {
            fmt.Printf("Successfully launched into background")
        }
        return
    }

    // some fancy "who am I splash" output :-)
    fmt.Printf( "\n%s - %s\nfor %s\n%s\n\n", ircbotint.IrcBotName, ircbotint.IrcBotVersion,
        ircbotint.IrcBotParentProject, ircbotint.IrcBotC )

    // set up main control channel for communication from all worker-threads
    mainCtrl = make(chan string, 1)

    // flag to keep all worker threads running or tell them to exit
    running = true

    for !shutdown {

        hcIrc = hcirc.New( "irc.hellcat.net", "6667", "Testuser", "Testnick", "" )
        //hcIrc = hcirc.New("192.168.241.10", "16667", "Testuser", "Testnick", "")
        hcIrc.Debugmode = cmdArgDebug
        hcIrc.Connect()
        if len(hcIrc.Error) == 0 {

            // fire up server messages queues
            hcIrc.StartInboundQueue()
            hcIrc.StartOutboundQueue()
            hcIrc.StartOutQuickQueue()

            // start main listener loop
            go serverListener(hcIrc)

            mainRunning = true
            for mainRunning {
                s = <- mainCtrl
                // TODO: handle messages from worker-threads
                s = s // silencing the compiler about currently unused variable
            }

            ////// TEST / SAMPLE CODE FOR WRITING TO AND READING FROM MESSAGE QUEUES //////
            //hcIrc.StartOutboundQueue()
            //hcIrc.StartInboundQueue()
            //hcIrc.OutboundQueue <- "JOIN #test"
            ////    for s := range hcIrc.InboundQueue {
            //running := true
            //var s string
            //for running {
            //    s = <-hcIrc.InboundQueue
            //    _, _, _, _, _, text := hcIrc.ParseMessage(s)
            //    fmt.Printf("%s\n", text)
            //    if text == "!bye" {
            //        fmt.Printf("BYE RECEIVED!\n")
            //        //hcIrc.StopInboundQueue()
            //        //hcIrc.StopOutboundQueue()
            //        running = false
            //        shutdown = true
            //    }
            //    if text == "!test" {
            //        hcIrc.FloodThrottle = 5
            //        for i := 1; i <= 10; i++ {
            //            s = fmt.Sprintf("PRIVMSG #test :-- Test %d --", i)
            //            fmt.Printf("-- Test %d --\n", i)
            //            hcIrc.OutboundQueue <- s
            //        }
            //    }
            //    if hcIrc.Error == "EOF" {
            //        // seems like connection broke down, lets get out and the outer loop try to reconnect
            //        running = false
            //    }
            //}
            ////// TEST / SAMPLE CODE FOR WRITING TO AND READING FROM MESSAGE QUEUES //////

        }

        hcIrc.Shutdown()
        hcIrc = nil

        if !shutdown {
            time.Sleep(time.Duration(10) * time.Second)
        }

    }

}
