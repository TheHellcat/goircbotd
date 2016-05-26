package ircbotextsup

import (
    "hellcat/hcirc"
    "ircbotd/internal/ircbotint"
    "fmt"
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

    m["id"] = "4"
    m["data"] = "dong"
    ircbotint.DmSet("test", "toast", []string{"id"}, m)
    m["id"] = "5"
    m["data"] = "dong"
    ircbotint.DmSet("test", "toast", []string{"id"}, m)
    m["id"] = "6"
    m["data"] = "dong"
    ircbotint.DmSet("test", "toast", []string{"id"}, m)

    m = make(map[string]string)
    m["data"] = "dong"
    r, i := ircbotint.DmGet("test", "toast", []string{"id", "data"}, m)
    fmt.Println(i, r)
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
