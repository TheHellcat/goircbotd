package ircbotint

import (
    "hellcat/hcirc"
)

type ChatCommandCallback func(string, string, string, string, string, string, string) string

var regedChatCommandsInt map[string]ChatCommandCallback
var hcIrc *hcirc.HcIrc


/**
 *
 */
func InitChatcmdHan(pHcIrc *hcirc.HcIrc) {
    hcIrc = pHcIrc
    regedChatCommandsInt = make(map[string]ChatCommandCallback)
}


/**
 *
 */
func RegisterInternalChatCommand(command string, function ChatCommandCallback) {
    regedChatCommandsInt[command] = function
}


/**
 *
 */
func GetRegisteredInternalChatCommands() map[string]ChatCommandCallback {
    return regedChatCommandsInt
}


/**
 *
 */
func executeCommand(command, channel, nick, user, host, cmd, param string, function ChatCommandCallback) {
    var s string

    s = function(command, channel, nick, user, host, cmd, param)
    hcIrc.OutboundQueue <- s
}


/**
 *
 */
func HandleCommand(command, channel, nick, user, host, cmd, param string) {
    var function ChatCommandCallback
    var exists bool

    function, exists = regedChatCommandsInt[cmd]
    if exists {
        go executeCommand(command, channel, nick, user, host, cmd, param, function)
    }
}
