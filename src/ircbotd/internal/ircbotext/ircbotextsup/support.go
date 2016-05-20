package ircbotextsup

import (
    "fmt"
    "ircbotd/internal/ircbotint"
)

func helpChatCommand(command, channel, nick, user, host, cmd, param string) string {
    var s, t string
    var cmdList map[string]ircbotint.ChatCommandCallback

    cmdList = ircbotint.GetRegisteredInternalChatCommands()
    t = ""
    for s, _ = range cmdList {
        t = fmt.Sprintf("%s %s", t, s)
    }
    t = fmt.Sprintf("PRIVMSG %s :Bot-Commands: %s", channel, t)

    return t
}
