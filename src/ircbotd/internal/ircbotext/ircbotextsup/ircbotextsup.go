package ircbotextsup

import (
    "fmt"
    "strings"
    "hellcat/hcirc"
    "ircbotd/internal/ircbotint"
)

func debugPrint(hcIrc *hcirc.HcIrc, s string) {
    if hcIrc.Debugmode {
        s = fmt.Sprintf("[SUPPORTEXTENSIONDEBUG] %s", s)
        s = strings.Replace(s, string('\n'), "", -1)
        s = strings.Replace(s, string('\r'), "", -1)
        fmt.Printf("%s\n", s)
    }
}


/**
 * Initialise extension and register additional functionality with the main code
 */
func SupportExtensionInit(hcIrc *hcirc.HcIrc) {
    ircbotint.RegisterInternalChatCommand("!help", helpChatCommand)
    ircbotint.RegisterInternalChatCommand("!commands", helpChatCommand)
}


/**
 * Clean up and take care the bot can be shutdown or - more importantly - be restarted without issues
 */
func SupportExtensionShutdown(hcIrc *hcirc.HcIrc) {
}
