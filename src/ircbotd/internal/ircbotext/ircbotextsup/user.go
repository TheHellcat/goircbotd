package ircbotextsup

import (
    "hellcat/hcirc"
    "ircbotd/internal/ircbotint"
)


/**
 *
 */
func UsermanExtensionInit(hcIrc *hcirc.HcIrc) {
    //ircbotint.RegisterInternalChatCommand("!sample", sampleChatCommand)
    ircbotint.DmCheckTable("test", "toast", "CREATE TABLE `toast` (`id` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, `data` TEXT );")
    m := make(map[string]string)
    m["id"] = "1"
    m["data"] = "meep"
    ircbotint.DmSet("test", "toast", []string{"id"}, m)

    m["id"] = "2"
    m["data"] = "moop"
    ircbotint.DmSet("test", "toast", []string{"id"}, m)
    m["id"] = "2"
    m["data"] = "woop"
    ircbotint.DmSet("test", "toast", []string{"id"}, m)

    m["id"] = "3"
    m["data"] = "ding"
    ircbotint.DmSet("test", "toast", []string{"id"}, m)
}


/**
 *
 */
func UsermanExtensionShutdown(hcIrc *hcirc.HcIrc) {

}


/**
 *
 */
func opChatCommand(command, channel, nick, user, host, cmd, param string) string {
    return ""
}
