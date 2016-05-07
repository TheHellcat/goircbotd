package main

import
(
    "fmt"
    "flag"
    "hellcat/hcirc"
    "time"
)

var cmdArgDebug bool

func init() {
    flag.BoolVar(&cmdArgDebug, "debug", false, "Enable debug-mode")
}

func main() {

    var hcIrc *hcirc.HcIrc
    var shutdown bool = false

    flag.Parse()

    for !shutdown {

        //hcIrc = hcirc.New( "irc.hellcat.net", "6667", "Testuser", "Testnick", "" )
        hcIrc = hcirc.New("192.168.241.10", "16667", "Testuser", "Testnick", "")

        hcIrc.Debugmode = cmdArgDebug
        hcIrc.Connect()
        if len(hcIrc.Error) == 0 {

            ////// TEST / SAMPLE CODE FOR WRITING TO AND READING FROM MESSAGE QUEUES //////
            hcIrc.StartOutboundQueue()
            hcIrc.StartInboundQueue()
            hcIrc.OutboundQueue <- "JOIN #test"
            //    for s := range hcIrc.InboundQueue {
            running := true
            var s string
            for running {
                s = <-hcIrc.InboundQueue
                _, _, _, _, _, text := hcIrc.ParseMessage(s)
                fmt.Printf("%s\n", text)
                if text == "!bye" {
                    fmt.Printf("BYE RECEIVED!\n")
                    //hcIrc.StopInboundQueue()
                    //hcIrc.StopOutboundQueue()
                    running = false
                    shutdown = true
                }
                if text == "!test" {
                    hcIrc.FloodThrottle = 5
                    for i := 1; i <= 10; i++ {
                        s = fmt.Sprintf("PRIVMSG #test :-- Test %d --", i)
                        fmt.Printf("-- Test %d --\n", i)
                        hcIrc.OutboundQueue <- s
                    }
                }
                if hcIrc.Error == "EOF" {
                    // seems like connection broke down, lets get out and the outer loop try to reconnect
                    running = false
                }
            }
            ////// TEST / SAMPLE CODE FOR WRITING TO AND READING FROM MESSAGE QUEUES //////

        }

        hcIrc.Shutdown()
        hcIrc = nil

        if !shutdown {
            time.Sleep(time.Duration(10) * time.Second)
        }

    }

}
