package hcirc

import (
    "fmt"
    "strings"
)


func (hcIrc *HcIrc) conscmdSay( params string ) string {
    var a []string
    var s string
    var r string
    a = strings.SplitN( params, " ", 2 )
    if len(a) > 1 {
        s = fmt.Sprintf( "PRIVMSG %s :%s", a[0], a[1] )
        hcIrc.OutboundQueue <- s
        r = fmt.Sprintf( "Sent text to %s\n", a[0] )
    } else {
        r = fmt.Sprintf( "Not enough parameters, usage: say <#channel> <text>" )
    }

    return r
}


func (hcIrc *HcIrc) RegisterAdditionalConsoleCommands() {
    hcIrc.RegisterConsoleCommand( "say", "Says something into the given channel", hcIrc.conscmdSay )
}
