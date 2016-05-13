package hcirc

import (
    "fmt"
    "strings"
)


func (hcIrc *HcIrc) conscmdSay( command, params string ) string {
    var a []string
    var s string
    var r string
    a = strings.SplitN( params, " ", 2 )
    if len(a) > 1 {
        s = fmt.Sprintf( "PRIVMSG %s :%s", a[0], a[1] )
        hcIrc.OutboundQueue <- s
        r = fmt.Sprintf( "Sent text to %s\n", a[0] )
    } else {
        r = fmt.Sprintf( "Not enough parameters, usage: say <#channel> <text>\n" )
    }

    return r
}


func (hcIrc *HcIrc) conscmdChannels( command, params string ) string {
    return ""
}


func (hcIrc *HcIrc) conscmdJoin( command, params string ) string {
    var s string
    s = fmt.Sprintf( "JOIN %s", params )
    hcIrc.OutQuickQueue <- s
    return ""
}


func (hcIrc *HcIrc) conscmdPart( command, params string ) string {
    var s string
    s = fmt.Sprintf( "PART %s", params )
    hcIrc.OutQuickQueue <- s
    return ""
}


func (hcIrc *HcIrc) RegisterAdditionalConsoleCommands() {
    hcIrc.RegisterConsoleCommand( "say", "Says something into the given channel", hcIrc.conscmdSay )
    hcIrc.RegisterConsoleCommand( "channels", "Lists all channels the bot is joined into", hcIrc.conscmdChannels )
    hcIrc.RegisterConsoleCommand( "join", "Joins a channel", hcIrc.conscmdJoin )
    hcIrc.RegisterConsoleCommand( "part", "Parts a channel", hcIrc.conscmdPart )
}
