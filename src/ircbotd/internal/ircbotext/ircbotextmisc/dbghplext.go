package ircbotextmisc

import "hellcat/hcirc"


/**
 *
 */
func DbgHelpExtensionInit(hcIrc *hcirc.HcIrc) {
    if hcIrc.Debugmode {
        hcIrc.RegisterConsoleCommand("sim-srvmsg", "Simulates a message received by the IRC server", dbgSimSrvmsgConsoleCommand)
    }
}


/**
 *
 */
func DbgHelpExtensionShutdown(hcIrc *hcirc.HcIrc) {
    // NOP
}


/**
 *
 */
func dbgSimSrvmsgConsoleCommand(command, params string) string {
    var hcIrc *hcirc.HcIrc

    hcIrc = hcirc.Self

    hcIrc.InboundQueue <- params

    return ""
}
