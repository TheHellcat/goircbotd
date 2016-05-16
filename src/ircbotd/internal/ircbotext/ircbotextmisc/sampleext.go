package ircbotextmisc

import (
    "fmt"
    "strings"
    "hellcat/hcirc"
    "ircbotd/internal/ircbotint"
)

func debugPrint(hcIrc *hcirc.HcIrc, s string) {
    if hcIrc.Debugmode {
        s = fmt.Sprintf("[SAMPLEEXTENSIONDEBUG] %s", s)
        s = strings.Replace(s, string('\n'), "", -1)
        s = strings.Replace(s, string('\r'), "", -1)
        fmt.Printf("%s\n", s)
    }
}


/**
 * Initialise extension and register additional functionality with the main code
 */
func SampleExtensionInit(hcIrc *hcirc.HcIrc) {
    debugPrint(hcIrc, "Initialising sample extension")

    hcIrc.RegisterConsoleCommand("sample", "Sample console command from the sample extension", sampleConsoleCommand)
    ircbotint.RegisterInternalChatCommand("!sample", sampleChatCommand)
}


/**
 * Clean up and take care the bot can be shutdown or - more importantly - be restarted without issues
 */
func SampleExtensionShutdown(hcIrc *hcirc.HcIrc) {
    debugPrint(hcIrc, "Shutting down sample extension")
}

func sampleConsoleCommand(command, params string) string {
    return "This is a sample console command, it doesn't really do anything\nbut demonstrate how to add/register additional commands.\n"
}

func sampleChatCommand(command, channel, nick, user, host, cmd, param string) string {
    var s string
    s = fmt.Sprintf("PRIVMSG %s :%s: This is a sample chat command, it doesn't really do anything but demonstrate how to add/register additional commands.\n", channel, nick)
    return s
}
