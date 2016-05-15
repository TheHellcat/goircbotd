package ircbotextmisc

import (
    "fmt"
    "strings"
    "hellcat/hcirc"
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
}

/**
 * Clean up and take care the bot can be shutdown or - more importantly - be restarted without issues
 */
func SampleExtensionShutdown(hcIrc *hcirc.HcIrc) {
    debugPrint(hcIrc, "Shutting down sample extension")
}
