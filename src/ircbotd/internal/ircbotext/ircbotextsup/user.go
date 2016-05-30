package ircbotextsup

import (
    "fmt"
    "hellcat/hcirc"
    "ircbotd/internal/ircbotint"
    "strings"
    "time"
)


var umanMsgChan chan hcirc.ServerMessage


/**
 *
 */
func UsermanExtensionInit(hcIrc *hcirc.HcIrc) {
    ircbotint.DmCheckTable( "user", "autoop", "CREATE TABLE `autoop` ( `id` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE, `channel` TEXT, `nick` TEXT, `mode` TEXT, `setby` TEXT, `ts` INTEGER);" )
    ircbotint.RegisterInternalChatCommand( "!op", opChatCommand )

    umanMsgChan = make(chan hcirc.ServerMessage, hcIrc.QueueSize)
    hcIrc.RegisterServerMessageHook("usermanagerextension", umanMsgChan)

    //ircbotint.RegisterInternalChatCommand("!sample", sampleChatCommand)
    /*ircbotint.DmCheckTable("test", "toast", "CREATE TABLE `toast` (`id` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, `data` TEXT );")
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

    m = make(map[string]string)
    m["data"] = "zing%"
    ircbotint.DmDelete("test", "toast", m)*/
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
    var nickNormaled string
    var hcIrc *hcirc.HcIrc
    var chanUsers map[string]userinfo
    var uModes string
    var returnData string
    var dbData map[string]string

    returnData = ""
    dbData = make(map[string]string)

    hcIrc = hcirc.Self
    chanUsers = hcIrc.GetChannelUsers( channel )
    nickNormaled = hcIrc.NormalizeNick( nick )
    uModes = chanUsers[nickNormaled].NickModes
    if strings.Contains( uModes, "@" ) {
        returnData = fmt.Sprintf( "MODE %s +o %s", channel, param )
        dbData["channel"] = channel
        dbData["nick"] = hcIrc.NormalizeNick( param )
        dbData["mode"] = "o"
        dbData["setby"] = nick
        dbData["ts"] = string( time.Now().Unix() )
        ircbotint.DmSet( "user", "autoop", []string{"channel", "nick", "mode"}, dbData )
    }

    return returnData
}
