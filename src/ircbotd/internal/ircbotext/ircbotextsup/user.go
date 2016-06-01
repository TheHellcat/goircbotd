package ircbotextsup

import (
    "fmt"
    "strings"
    "time"
    "strconv"
    "hellcat/hcirc"
    "hellcat/hcthreadutils"
    "ircbotd/internal/ircbotint"
)


var umanMsgChan chan hcirc.ServerMessage
var umanUserCache map[string]string
var umanJoinHanRunning bool
var umanJoinHanThreadId string


/**
 *
 */
func UsermanExtensionInit(hcIrc *hcirc.HcIrc) {
    umanUserCache = make(map[string]string)

    ircbotint.DmCheckTable( "user", "autoop", "CREATE TABLE `autoop` ( `id` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE, `channel` TEXT, `nick` TEXT, `mode` TEXT, `setby` TEXT, `ts` INTEGER);" )
    ircbotint.RegisterInternalChatCommand( "!op", opChatCommand )
    ircbotint.RegisterInternalChatCommand( "!deop", deopChatCommand )
    ircbotint.RegisterInternalChatCommand( "!voice", voiceChatCommand )
    ircbotint.RegisterInternalChatCommand( "!devoice", devoiceChatCommand )

    umanMsgChan = make(chan hcirc.ServerMessage, hcIrc.QueueSize)
    hcIrc.RegisterServerMessageHook("usermanagerextension", umanMsgChan)

    umanJoinHanRunning = true
    go umanJoinHandler()
}


/**
 *
 */
func UsermanExtensionShutdown(hcIrc *hcirc.HcIrc) {
    umanJoinHanRunning = false
    close(umanMsgChan)
    hcthreadutils.WaitForRoutinesEndById( []string{ umanJoinHanThreadId } )
    umanMsgChan = nil
}


/**
 *
 */
func umanJoinHandler() {
    var msg hcirc.ServerMessage
    var hcIrc *hcirc.HcIrc
    var exists bool
    var nick string
    var s string
    var mode string
    var rs map[int]map[string]string
    var rv map[string]string
    var m map[string]string
    var i int

    hcIrc = hcirc.Self

    if hcIrc.Debugmode {
        fmt.Printf( "[USERMANEXTDEBUG] ON-JOIN handler thread starting\n" )
    }
    umanJoinHanRunning = true

    umanJoinHanThreadId = hcthreadutils.GetRoutineId()

    for msg = range umanMsgChan {
        if "JOIN" == msg.Command {
            nick = fmt.Sprintf( "%s-%s", hcIrc.NormalizeNick( msg.Channel ), hcIrc.NormalizeNick( msg.Nick ))
            mode, exists = umanUserCache[nick]
            if !exists {
                m = make(map[string]string)  // lets hope GC clears up the old one....
                m["channel"] = msg.Channel
                m["nick"] = hcIrc.NormalizeNick( msg.Nick )
                if hcIrc.Debugmode {
                    fmt.Printf( "[USERMANEXTDEBUG] Looking up auto-mode setting for '%s' in '%s'\n", m["nick"], m["channel"] )
                }
                rs, i = ircbotint.DmGet( "user", "autoop", []string{"mode"}, m )
                if i > 0 {
                    rv = rs[0]
                    mode = rv["mode"]
                    if hcIrc.Debugmode {
                        fmt.Printf( "[USERMANEXTDEBUG] Found auto-mode '%s' for '%s'\n", mode, nick )
                    }
                } else {
                    mode = ""
                    if hcIrc.Debugmode {
                        fmt.Printf( "[USERMANEXTDEBUG] No auto-mode found for '%s'\n", nick )
                    }
                }
                umanUserCache[nick] = mode
            } else {
                if hcIrc.Debugmode {
                    fmt.Printf( "[USERMANEXTDEBUG] Using cached data for '%s' in '%s'\n", m["nick"], m["channel"] )
                }
            }
            if len(mode) > 0 {
                if hcIrc.Debugmode {
                    fmt.Printf( "[USERMANEXTDEBUG] Setting mode for '%s' to '%s'\n", nick, mode )
                }
                s = fmt.Sprintf( "MODE %s +%s %s", msg.Channel, mode, msg.Nick )
                hcIrc.OutQuickQueue <- s
            } else {
                if hcIrc.Debugmode {
                    fmt.Printf( "[USERMANEXTDEBUG] Not setting mode for '%s'\n", nick )
                }
            }
        }
    }

    if hcIrc.Debugmode {
        fmt.Printf( "[USERMANEXTDEBUG] ON-JOIN handler thread terminating\n" )
    }
    umanJoinHanRunning = false
}


/**
 *
 */
func umanIsOp( channel, nick string ) bool {
    var nickNormaled string
    var hcIrc *hcirc.HcIrc
    var chanUsers map[string]hcirc.Userinfo
    var uModes string
    var returnData bool

    returnData = false

    hcIrc = hcirc.Self
    chanUsers = hcIrc.GetChannelUsers( channel )
    nickNormaled = hcIrc.NormalizeNick( nick )
    uModes = chanUsers[nickNormaled].NickModes
    if strings.Contains( uModes, "@" ) {
        returnData = true
    }

    return returnData
}


/**
 *
 */
func modeChatCommand(command, channel, nick, user, host, cmd, param, mode string) string {
    var hcIrc *hcirc.HcIrc
    var returnData string
    var dbData map[string]string
    var s string

    returnData = ""
    dbData = make(map[string]string)

    if umanIsOp( channel, nick ) {
        returnData = fmt.Sprintf( "MODE %s +%s %s", channel, mode, param )
        dbData["channel"] = channel
        dbData["nick"] = hcIrc.NormalizeNick( param )
        dbData["mode"] = mode
        dbData["setby"] = nick
        dbData["ts"] = strconv.FormatInt( time.Now().Unix(), 10 )
        ircbotint.DmSet( "user", "autoop", []string{"channel", "nick"}, dbData )
    }

    // invalidate cache
    s = fmt.Sprintf( "%s-%s", hcIrc.NormalizeNick( channel ), hcIrc.NormalizeNick( param ))
    delete(umanUserCache, s)

    return returnData
}


/**
 *
 */
func unmodeChatCommand(command, channel, nick, user, host, cmd, param, mode string) string {
    var hcIrc *hcirc.HcIrc
    var returnData string
    var dbData map[string]string
    var s string

    returnData = ""
    dbData = make(map[string]string)

    if umanIsOp( channel, nick ) {
        returnData = fmt.Sprintf( "MODE %s -%s %s", channel, mode, param )
        dbData["channel"] = channel
        dbData["nick"] = hcIrc.NormalizeNick( param )
        dbData["mode"] = mode
        ircbotint.DmDelete( "user", "autoop", dbData )
    }

    // invalidate cache
    s = fmt.Sprintf( "%s-%s", hcIrc.NormalizeNick( channel ), hcIrc.NormalizeNick( param ))
    delete(umanUserCache, s)

    return returnData
}


/**
 *
 */
func opChatCommand(command, channel, nick, user, host, cmd, param string) string {
    return modeChatCommand(command, channel, nick, user, host, cmd, param, "o")
}


/**
 *
 */
func deopChatCommand(command, channel, nick, user, host, cmd, param string) string {
    return unmodeChatCommand(command, channel, nick, user, host, cmd, param, "o")
}


/**
 *
 */
func voiceChatCommand(command, channel, nick, user, host, cmd, param string) string {
    return modeChatCommand(command, channel, nick, user, host, cmd, param, "v")
}


/**
 *
 */
func devoiceChatCommand(command, channel, nick, user, host, cmd, param string) string {
    return unmodeChatCommand(command, channel, nick, user, host, cmd, param, "v")
}
