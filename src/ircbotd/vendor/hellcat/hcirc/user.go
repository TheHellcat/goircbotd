package hcirc

import (
    "strings"
    "fmt"
    "regexp"
)


/**
 *
 */
func (hcIrc *HcIrc) AddUserToChannel(channel, nick string) {
    hcIrc.channelUserJoin(channel, nick)
}


/**
 *
 */
func (hcIrc *HcIrc) RemoveUserFromChannel(channel, nick string) {
    hcIrc.channelUserPart(channel, nick)
}


/**
 *
 */
func (hcIrc *HcIrc) GetChannelUsers(channel string) map[string]userinfo {
    var uList userlist
    var uInfoList map[string]userinfo
    var exists bool
    var nick, displayname string
    var user userinfo

    uInfoList = make(map[string]userinfo)

    uList, exists = hcIrc.channelUsers[channel]

    if exists {
        // put the users into a nice map with all the details we have
        for nick, displayname = range uList {
            user.NickDislpayname = hcIrc.stripUsermodeChars(displayname)
            user.NickModes = hcIrc.getUsermodeChars(displayname)
            user.NickNormalizedName = nick
            uInfoList[nick] = user
        }
    }

    return uInfoList
}


/**
 *
 */
func (hcIrc *HcIrc) NormalizeNick(nick string) string {
    return hcIrc.stripUsermodeChars(strings.ToLower(nick))
}


/**
 * Adds a user to a channels userlist
 */
func (hcIrc *HcIrc) channelUserJoin(channel, nick string) {
    var s, t string
    var uList userlist
    var exists bool

    uList, exists = hcIrc.channelUsers[channel]
    if !exists {
        uList = make(userlist)
    }
    s = hcIrc.NormalizeNick(nick)

    // check if user is already in the list and add, if not
    _, exists = uList[s]
    if !exists {
        uList[s] = nick
    } else {
        t = hcIrc.getUsermodeChars( nick )
        if len(t) > 0 {
            // nickname contains usermode(s), so let's use it anyways as they might be more recent/correct
            // as what we know about at the moment
            uList[s] = nick
        }
    }

    hcIrc.channelUsers[channel] = uList

    // check if we ourselves just joined that room and remember we're in here if so
    if s == hcIrc.NormalizeNick(hcIrc.nick) {
        // jap, it's us
        hcIrc.JoinedChannels[channel] = channel
        hcIrc.debugPrint("Joined channel:", channel)
    }

    if hcIrc.Debugmode {
        if !exists {
            t := ""
            for _, u := range uList {
                t = fmt.Sprintf("%s,%s", t, u)
            }
            t = strings.Trim(t, ",")
            s = fmt.Sprintf("Updated userlist after JOIN for channel %s:", channel)
            hcIrc.debugPrint(s, t)
        } else {
            if len(t) == 0 {
                s = fmt.Sprintf("Skipped userlist update after JOIN (user already in list) for channel %s", channel)
                hcIrc.debugPrint(s, "")
            } else {
                s = fmt.Sprintf("Updated '%s' in userlist after JOIN (was already in list but new nick had modes) for channel %s", nick, channel)
                hcIrc.debugPrint(s, "")
            }
        }
    }
}


/**
 * Removes a user from a channels userlist
 */
func (hcIrc *HcIrc) channelUserPart(channel, nick string) {
    var uList userlist
    var exists bool

    uList, exists = hcIrc.channelUsers[channel]
    if exists {
        delete(uList, hcIrc.NormalizeNick(nick))
        hcIrc.channelUsers[channel] = uList
    }

    if hcIrc.Debugmode {
        t := ""
        for _, u := range uList {
            t = fmt.Sprintf("%s,%s", t, u)
        }
        t = strings.Trim(t, ",")
        s := fmt.Sprintf("Updated userlist after PART for channel %s:", channel)
        hcIrc.debugPrint(s, t)
    }
}


/**
 *
 */
func (hcIrc *HcIrc) changeUserMode(channel, nick, raw string) {
    var normalUsername string
    var savedNick string
    var newNick string
    var a []string
    var i, j, n int
    var s string
    var modeCommands string
    var modeCommand string
    var nicks []string
    var uList userlist
    var exists bool
    var sign string

    if channel == hcIrc.nick {
        // if the "channel name" is actually our own nickname, the server just told us about our global
        // user modes

        a = strings.Split(raw, ":")
        s = fmt.Sprintf("Got own global usermodes: %s", a[1])
        hcIrc.debugPrint(s, "")
        // not doing much with that information at the moment
        return
    }

    a = strings.SplitN(raw, " ", 4)
    i = len(a)
    nicks = strings.Split(a[i - 1], " ")
    modeCommands = a[i - 2]

    sign = "+"
    n = 0
    i = len(modeCommands)
    for j = 0; j < i; j++ {
        s = modeCommands[j:j+1]
        if "+" == s || "-" == s {
            sign = s
        } else {
            modeCommand = fmt.Sprintf( "%s%s", sign, s )
            normalUsername = hcIrc.NormalizeNick(nicks[n])

            s = fmt.Sprintf("Got mode change for nick '%s' on channel '%s':", normalUsername, channel)
            hcIrc.debugPrint(s, modeCommand)

            uList = hcIrc.channelUsers[channel]
            savedNick, exists = uList[normalUsername]
            if !exists {
                hcIrc.debugPrint("Unable to change mode: user not found in channel.", "")
            } else {
                newNick = savedNick

                if "-q" == modeCommand {
                    newNick = strings.Replace(savedNick, "~", "", -1)
                } else if "+q" == modeCommand {
                    newNick = fmt.Sprintf("~%s", savedNick)
                } else if "-a" == modeCommand {
                    newNick = strings.Replace(savedNick, "&", "", -1)
                } else if "+a" == modeCommand {
                    newNick = fmt.Sprintf("&%s", savedNick)
                } else if "-o" == modeCommand {
                    newNick = strings.Replace(savedNick, "@", "", -1)
                } else if "+o" == modeCommand {
                    newNick = fmt.Sprintf("@%s", savedNick)
                } else if "-h" == modeCommand {
                    newNick = strings.Replace(savedNick, "%", "", -1)
                } else if "+h" == modeCommand {
                    newNick = fmt.Sprintf("%s%s", "%", savedNick)
                } else if "-v" == modeCommand {
                    newNick = strings.Replace(savedNick, "+", "", -1)
                } else if "+v" == modeCommand {
                    newNick = fmt.Sprintf("+%s", savedNick)
                }

                s = fmt.Sprintf("Changed modes: from '%s' to '%s'", savedNick, newNick)
                hcIrc.debugPrint(s, "")
                uList[normalUsername] = newNick
                hcIrc.channelUsers[channel] = uList
            }

            n++
        }
    }
}


/**
 * Strips all non alpha-numeric chars off the username (like mode indicators @, +, etc.)
 */
func (hcIrc *HcIrc) stripUsermodeChars(nick string) string {
    var err error
    var regex *regexp.Regexp
    var s string

    regex, err = regexp.Compile("[^A-Za-z0-9]")
    if err == nil {
        s = regex.ReplaceAllString(nick, "")
    } else {
        s = ""
    }

    return s
}


/**
 *
 */
func (hcIrc *HcIrc) getUsermodeChars(nick string) string {
    var err error
    var regex *regexp.Regexp
    var s string

    // user mode indicators:
    // ~ - some IRCs: channel owner
    // & - some IRCs: channel admin
    // @ - channel OP (mod for Twitch)
    // % - channel HOP (half-OP)
    // + - has voice
    // ! - own (for Twitch support): broadcaster
    // $ - own (for Twitch support): subscriber
    // / - own (for Twitch support): follower
    // * - own (for Twitch support): has Turbo
    regex, err = regexp.Compile("[^~&@%+!$\\/*]")
    if err == nil {
        s = regex.ReplaceAllString(nick, "")
    } else {
        s = ""
    }

    return s
}
