package ircbotext

import (
    "hellcat/hcirc"
    "ircbotd/internal/ircbotext/ircbotextmisc"
    "ircbotd/internal/ircbotext/ircbotextsup"
)


/**
 * Initialize all extensions
 *
 * This function calls the setup/init functions of all extensions.
 * For every extension that is supposed to be available, one call to it's setup/init function
 * needs to be added here.
 */
func InitExtensions(hcIrc *hcirc.HcIrc) {
    ircbotextmisc.SampleExtensionInit(hcIrc)
    ircbotextsup.SupportExtensionInit(hcIrc)
}


/**
 * Shutdown all extensions
 *
 * This function calls the shutdown functions of all extensions.
 * For every extension that is active, one call to it's setup/init function
 * needs to be added here - unless no special shutdown operations need to be performed.
 */
func ShutdownExtensions(hcIrc *hcirc.HcIrc) {
    ircbotextmisc.SampleExtensionShutdown(hcIrc)
    ircbotextsup.SupportExtensionShutdown(hcIrc)
}
