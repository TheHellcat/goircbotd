package hcirc

import (
    "fmt"
    "strings"
)

func (hcIrc *HcIrc) conscmdSay(command, params string) string {
    var a []string
    var s string
    var r string
    a = strings.SplitN(params, " ", 2)
    if len(a) > 1 {
        s = fmt.Sprintf("PRIVMSG %s :%s", a[0], a[1])
        hcIrc.OutboundQueue <- s
        r = fmt.Sprintf("Sent text to %s\n", a[0])
    } else {
        r = fmt.Sprintf("Not enough parameters, usage: say <#channel> <text>\n")
    }

    return r
}

func (hcIrc *HcIrc) conscmdChannels(command, params string) string {
    var s string
    var out string

    out = ""
    for _, s = range hcIrc.JoinedChannels {
        out = fmt.Sprintf("%s%s\n", out, s)
    }
    out = fmt.Sprintf("Joined channels:\n\n%s", out)

    return out
}

func (hcIrc *HcIrc) conscmdUsers(command, params string) string {
    var channel string
    var users userlist
    var out string
    var user string

    out = ""
    for channel, users = range hcIrc.channelUsers {
        out = fmt.Sprintf("%s\n%s:\n", out, channel)
        for _, user = range users {
            out = fmt.Sprintf("%s  %s\n", out, user)
        }
    }
    out = fmt.Sprintf("Users per channel:\n%s", out)

    return out
}

func (hcIrc *HcIrc) conscmdJoin(command, params string) string {
    var s string
    s = fmt.Sprintf("JOIN %s", params)
    hcIrc.OutQuickQueue <- s
    return ""
}

func (hcIrc *HcIrc) conscmdPart(command, params string) string {
    var s string
    s = fmt.Sprintf("PART %s", params)
    hcIrc.OutQuickQueue <- s
    return ""
}

func (hcIrc *HcIrc) RegisterAdditionalConsoleCommands() {
    hcIrc.RegisterConsoleCommand("say", "Says something into the given channel", hcIrc.conscmdSay)
    hcIrc.RegisterConsoleCommand("channels", "Lists all channels the bot is joined into", hcIrc.conscmdChannels)
    hcIrc.RegisterConsoleCommand("join", "Joins a channel", hcIrc.conscmdJoin)
    hcIrc.RegisterConsoleCommand("part", "Parts a channel", hcIrc.conscmdPart)
    hcIrc.RegisterConsoleCommand("users", "Lists all online users in all joined", hcIrc.conscmdUsers)
}
