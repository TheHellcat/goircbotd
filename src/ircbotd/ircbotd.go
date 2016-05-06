package main

import
(
    "fmt"
    "flag"
    "hellcat/hcirc"
)


var cmdArgDebug bool


func init() {
    flag.BoolVar( &cmdArgDebug, "debug", false, "Enable debug-mode" )
}


func main() {

    var hcIrc *hcirc.HcIrc

    flag.Parse()

    hcIrc = hcirc.New( "irc.hellcat.net", "6667", "Testuser", "Testnick", "" )

    hcIrc.Debugmode = cmdArgDebug
    hcIrc.Connect()

    hcIrc.StartOutboundQueue()
    hcIrc.StartInboundQueue()
    hcIrc.OutboundQueue <- "JOIN #test"
//    for s := range hcIrc.InboundQueue {
    running := true
    var s string
    for running {
        s = <- hcIrc.InboundQueue
        _, _, _, _, _, text := hcIrc.ParseMessage(s)
        fmt.Printf("%s\n", text)
        if text == "!bye" {
            fmt.Printf("BYE RECEIVED!\n")
            hcIrc.StopInboundQueue()
            hcIrc.StopOutboundQueue()
            running = false
        }
        if text == "!test" {
            hcIrc.FloodThrottle = 10
            for i := 1; i <= 10; i++ {
                s = fmt.Sprintf( "PRIVMSG #test :-- Test %d --", i )
                fmt.Printf( "-- Test %d --\n", i )
                hcIrc.OutboundQueue <- s
            }
        }
    }

}
