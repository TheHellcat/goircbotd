package ircbotint

import (
    "gorilla/websocket"
    "net/http"
    "fmt"
    "hellcat/hcirc"
    "time"
    "encoding/json"
    "strings"
    "hellcat/hcthreadutils"
    "html"
)

type wchatBufMsg struct {
    channel string
    nick    string
    nickId  string
    message string
    tags    string
}

var wchatBuf []wchatBufMsg
var wchatBufCur int
var wchatBufSize int
var wchatBufThId string
var wchatBufChId string
var wchatMsgChan chan hcirc.ServerMessage


/**
 *
 */
func webchatHistoryBuffer() {
    var msg hcirc.ServerMessage
    var i int

    wchatBufThId = hcthreadutils.GetRoutineId()
    wchatBufChId = fmt.Sprintf("webchatHistoryBuffer-%s", wchatBufThId)

    wchatBufSize = 100
    wchatBufCur = 0
    wchatBuf = make([]wchatBufMsg, wchatBufSize)

    if wsHcIrc.Debugmode {
        fmt.Printf("[WSCHATDEBUG] History buffering thread started (TID:%s)\n", wchatBufThId)
    }

    wchatMsgChan = make(chan hcirc.ServerMessage, wsHcIrc.QueueSize)
    wsHcIrc.RegisterServerMessageHook(wchatBufChId, wchatMsgChan)

    for msg = range wchatMsgChan {
        if "PRIVMSG" == msg.Command {
            if ( hcIrc.IsTwitchModeEnabled() ) {
                wchatBuf[wchatBufCur].nick = msg.Nick
            } else {
                wchatBuf[wchatBufCur].nick = fmt.Sprintf("%s%s", msg.NickMode, msg.Nick)
            }

            wchatBuf[wchatBufCur].channel = msg.Channel
            wchatBuf[wchatBufCur].nickId = wsHcIrc.NormalizeNick(msg.Nick)
            wchatBuf[wchatBufCur].message = msg.Text
            wchatBuf[wchatBufCur].tags = msg.Tags

            wchatBufCur++
            if wchatBufCur == wchatBufSize {
                wchatBufCur = 0
            }
        } else if "CLEARCHAT" == msg.Command {
            for i = 0; i < wchatBufSize; i++ {
                if wsHcIrc.NormalizeNick(msg.Text) == wchatBuf[i].nickId {
                    wchatBuf[i].message = ""
                }
            }
        } else if "JOIN" == msg.Command {
        } else if "PART" == msg.Command {
        }
    }

    if wsHcIrc.Debugmode {
        fmt.Printf("[WSCHATDEBUG] History buffering thread terminating (TID:%s)\n", wchatBufThId)
    }
}


/**
 *
 */
func initWebchat() {
    go webchatHistoryBuffer()
}


/**
 *
 */
func shutdownWebchat() {
    wsHcIrc.UnregisterServerMessageHook(wchatBufChId)
    close(wchatMsgChan)
    if len(wchatBufThId) > 0 {
        hcthreadutils.WaitForRoutinesEndById([]string{wchatBufThId})
    }
    wchatBufThId = ""
}


/**
 * client receiver thrread
 *
 * waits for messages being sent from the webclient and sends them to the request handler
 */
func webchatClientReceiver(conn *websocket.Conn, inChan chan string) {
    var mt int
    var ba []byte
    var message string
    var err error
    var running bool

    running = true

    if wsHcIrc.Debugmode {
        fmt.Printf("[WSCHATDEBUG] Receiver thread started\n")
    }

    for running {
        mt, ba, err = conn.ReadMessage()
        if err != nil {
            if wsHcIrc.Debugmode {
                fmt.Printf("[WSCHATDEBUG] Feiled to read from connection: %s\n", err.Error())
            }
            inChan <- "QUIT"
            running = false
        } else {
            message = string(ba)
            if mt == websocket.TextMessage {
                inChan <- message
            }
        }
    }

    if wsHcIrc.Debugmode {
        fmt.Printf("[WSCHATDEBUG] Receiver thread ended\n")
    }
}


/**
 *
 */
func generateWebchatJSON(text, id, nick, nickId, tags string) []byte {
    var msgType string
    var msgCss string
    var clientMsg map[string]string
    var ba []byte
    var err error
    var tagList map[string]string
    var emotes map[int]hcirc.TwitchMsgEmoteInfo
    var emoteCount int
    var i int
    var s, t string
    var badges map[int]hcirc.TwitchBadgeType
    var badgeCount int
    var nickColor string
    var styleOverride string
    var orgNick string
    var emoteCache map[string]string

    clientMsg = make(map[string]string)
    orgNick = nick
    emoteCache = make(map[string]string)

    // check if this is a regular message or a /me action
    msgType = "chatmessage"
    msgCss = "Text"
    if len(text) > 7 {
        if "\x01ACTION" == text[0:7] {
            msgType = "chataction"
            msgCss = "Action"
            text = strings.Replace(text, "\x01ACTION ", "", -1)
            text = strings.Replace(text, "\x01", "", -1)
        }
    }

    if ( hcIrc.IsTwitchModeEnabled() ) {
        tagList = hcIrc.ParseTwitchTags(tags)
        emotes, emoteCount = hcIrc.ParseTwitchEmoteTag(tagList["emotes"])
        badges, badgeCount = hcIrc.ParseTwitchBadgesTag(tagList["badges"], tagList["room-id"])
        nickColor = tagList["color"]
        styleOverride = ""
        if ( len(tagList["display-name"]) > 0) {
            nick = tagList["display-name"]
        }

        // Emotes Part 1 of 2:
        // we can't put in the actual emote images right here, as we do so by HMTL image tags
        // but as we're HTML escaping the whole message later on (to prevent XSS and similar things
        // in the webchat client) it would break the image tags.
        // So we first put in some HTML safe placeholders that - after the HTML escaping - will
        // then be string-replaced with the actual emote image tags (that we're also generating already and
        // saving into a temp map here)

        // cast the message string into a UTF8 safe rune slice, to have UTF8 chars not mess up everything
        // (UTF8 chars need more than 1 byte so it'd mess up the offsets of the emotes)
        r := []rune(text)

        // if there's no emotes in the message, "s" will never be filled with the placeholder'ed text,
        // so set it to the original message text or else we'd loose the message after the loop
        s = text
        for i = 0; i < emoteCount; i++ {
            t = fmt.Sprintf("{{%s}}", string(r[emotes[i].From:emotes[i].To + 1]))
            emoteCache[t] = fmt.Sprintf("<img src=\"%s\" alt=\"\" class=\"chatEmote\" />", emotes[i].ChatUrl)
            s = fmt.Sprintf("%s%s%s", string(r[:emotes[i].From]), t, string(r[emotes[i].To + 1:]))
            // cast our working string back to UFT8 safe rune slice as our string operations above return a
            // UTF8 not-safe string....
            r = []rune(s)
        }

        // set current message text to the emote-placeholder'ed version
        text = s

        // HTML escape the nickname....
        nick = html.EscapeString(nick)

        // ....and add the badges to it
        s = ""
        for i = 0; i < badgeCount; i++ {
            if ( len(badges[i].ImageUrl) > 0 ) {
                s = fmt.Sprintf("%s<img src=\"%s\" alt=\"%s\" class=\"chatUserBadge\" /> ", s, badges[i].ImageUrl, badges[i].Title)
            }
        }
        s = fmt.Sprintf("%s%s", s, nick)
        nick = s

        // nick custom color
        if ( len(nickColor) > 0) {
            styleOverride = fmt.Sprintf("%s color:%s;", styleOverride, nickColor)
        }
        styleOverride = strings.Trim(styleOverride, " ")
    }

    // no HTML allowed in webchat textmessages, replace "dangerous" characters with their respective entity
    text = html.EscapeString(text)

    // Emotes Part 2 of 2:
    // now that the message is safe (potential user HTML (breaking) escaped)
    // finally put the image-tags for the inline Twitch emotes in
    if ( hcIrc.IsTwitchModeEnabled() ) {
        for s, t = range emoteCache {
            // since all message text is HTML escaped at this point we need to escape our placeholder as well,
            // so if it includes HTML chars (like the "<" in the "<3" emote) the replacement still matches
            text = strings.Replace(text, html.EscapeString(s), t, -1)
        }
    }

    if ( "sys" == id[:3] ) {
        msgType = "sysevent"
        msgCss = "Event"
        styleOverride = "";
    }

    // if no nickId was supplied, generate one
    if len(nickId) < 1 {
        nickId = wsHcIrc.NormalizeNick(orgNick)
    }

    // build final JSON to be sent to web client
    clientMsg["type"] = msgType
    clientMsg["id"] = id
    clientMsg["cssClass"] = msgCss
    clientMsg["nick"] = nick
    clientMsg["nickId"] = nickId
    clientMsg["text"] = text
    clientMsg["styleOverride"] = styleOverride
    ba, err = json.Marshal(clientMsg)
    if err != nil {
        if wsHcIrc.Debugmode {
            fmt.Printf("[WSCHATDEBUG][HISTORYREPLAY] ERROR encoding JSON for client: %s\n", err.Error())
        }
    }

    return ba
}


/**
 * Check if the given text matches any of the set filter words.
 * Return also "true" if no filter words are set.
 */
func checkFilter(partialWordMap map[string]string, fullWordMap map[string]string, text string) bool {
    var match bool
    var s, t string

    match = false

    text = strings.ToLower(text)

    if ( (len(partialWordMap) > 0) || (len(fullWordMap) > 0) ) {
        for _, s = range partialWordMap {
            match = match || strings.Contains(text, s)
        }
        for _, s = range fullWordMap {
            t = " " + s + " "
            match = match || strings.Contains(text, t)
        }
    } else {
        match = true
    }

    return match
}


/**
 * main request handler
 *
 * handles the HTTP request, upgrade to WEBSOCKET and communication between bot and webclient
 */
func webchatHandler(writer http.ResponseWriter, request *http.Request) {
    var s string
    var msgChan chan hcirc.ServerMessage
    var inChan chan string
    var conn *websocket.Conn
    var err error
    var running bool
    var command, channel, nick, text string
    var myId string
    var srvMsg hcirc.ServerMessage
    var clientMsg map[string]string
    var clmsgid int
    var ba []byte
    var joinedChannels map[string]string
    var exists bool
    var a []string
    var msgChanId string
    var i int
    var filterPartialWords map[string]string
    var filterFullWords map[string]string

    running = true
    clmsgid = 0
    myId = request.RemoteAddr
    clientMsg = make(map[string]string)
    joinedChannels = make(map[string]string)
    filterPartialWords = make(map[string]string)
    filterFullWords = make(map[string]string)

    if wsHcIrc.Debugmode {
        fmt.Printf("[WSCHATDEBUG] New connection handler spawned: %s\n", myId)
    }

    // setup new channel to receive IRC server messages
    msgChan = make(chan hcirc.ServerMessage, wsHcIrc.QueueSize)
    // we need a unique ID for registering our channel
    s = fmt.Sprintf("webchat-%s-%d", request.RemoteAddr, time.Now().Unix())
    wsHcIrc.RegisterServerMessageHook(s, msgChan)
    if wsHcIrc.Debugmode {
        fmt.Printf("[WSCHATDEBUG] Registered server-messages channel with ID %s\n", s)
    }
    msgChanId = s

    // set up channel for receiving messages from webchat client
    inChan = make(chan string, wsHcIrc.QueueSize)

    // upgrade the HTTP connection to WEBSOCKETS
    conn, err = wsUpgrader.Upgrade(writer, request, nil)
    if err != nil {
        if wsHcIrc.Debugmode {
            fmt.Printf("[WSCHATDEBUG] Upgrading HTTP to WEBSOCKETS feiled: %s\n", err.Error())
        }
        return
    }
    defer conn.Close()

    // fork out the reader as separate routine/thread and listen on a chan
    // for it, this way we can have the read non-blocking and react on other
    // things as well while waiting for the client to send something
    go webchatClientReceiver(conn, inChan)

    for running {
        select {
        case s = <-inChan:
            a = strings.Split(s, " ")
            if "JOIN" == a[0] {
                joinedChannels[a[1]] = a[1]
                if wsHcIrc.Debugmode {
                    fmt.Printf("[WSCHATDEBUG] %s subscribed to channel %s\n", myId, a[1])
                }
                i = wchatBufCur + 2

                if ( (len(filterPartialWords) == 0) && (len(filterFullWords) == 0) ) {
                    if wsHcIrc.Debugmode {
                        fmt.Printf("[WSCHATDEBUG][HISTORYREPLAY] Sending buffer to client for channel %s\n", a[1])
                    }
                    for i != wchatBufCur + 1 {
                        if a[1] == wchatBuf[i].channel && len(wchatBuf[i].message) > 0 {
                            s = fmt.Sprintf("hist%d", i)

                            ba = generateWebchatJSON(wchatBuf[i].message, s, wchatBuf[i].nick, wchatBuf[i].nickId, wchatBuf[i].tags)

                            _ = conn.WriteMessage(websocket.TextMessage, ba)
                        }
                        i++
                        if i == wchatBufSize {
                            i = 0
                        }
                    }
                    if wsHcIrc.Debugmode {
                        fmt.Printf("[WSCHATDEBUG][HISTORYREPLAY] Done sending buffer to client\n")
                    }
                } else {
                    fmt.Printf("[WSCHATDEBUG][HISTORYREPLAY] Skipping history sending because filter are active\n")
                }

            } else if "PART" == a[0] {
                delete(joinedChannels, a[1])
                if wsHcIrc.Debugmode {
                    fmt.Printf("[WSCHATDEBUG] %s unsubscribed to channel %s\n", myId, a[1])
                }
            } else if "QUIT" == a[0] {
                running = false
            } else if "FILTERPARTIAL" == a[0] {
                filterPartialWords[a[1]] = strings.ToLower(a[1])
                if wsHcIrc.Debugmode {
                    fmt.Printf("[WSCHATDEBUG] %s added '%s' to partial words filter\n", myId, a[1])
                }
            } else if "FILTERFULL" == a[0] {
                filterFullWords[a[1]] = strings.ToLower(a[1])
                if wsHcIrc.Debugmode {
                    fmt.Printf("[WSCHATDEBUG] %s added '%s' to full words filter\n", myId, a[1])
                }
            }

        case srvMsg = <-msgChan:
            command = srvMsg.Command
            channel = srvMsg.Channel
            nick = srvMsg.Nick
            text = srvMsg.Text
            clmsgid += 1
            if clmsgid > 268435455 {
                // some cheap int32 kaboom protection by intentionally rolling over
                // this way there is a clearly defined behaviour when we come close to the limit
                // and no, I don't wanna use an int64 - it's also not really required here.
                clmsgid = 1
            }
            if "PRIVMSG" == command {
                _, exists = joinedChannels[channel]
                if (exists && checkFilter(filterPartialWords, filterFullWords, text) ) {
                    s = fmt.Sprintf("msg%d", clmsgid)
                    ba = generateWebchatJSON(text, s, nick, "", srvMsg.Tags)

                    err = conn.WriteMessage(websocket.TextMessage, ba)
                    if err != nil {
                        if wsHcIrc.Debugmode {
                            fmt.Printf("[WSCHATDEBUG] ERROR sending JSON to client: %s\n", err.Error())
                        }
                    }
                }
            } else if "JOIN" == command {
                s = fmt.Sprintf("$.%s", channel)
                _, exists = joinedChannels[s]
                if exists {
                    s = fmt.Sprintf("sys%d", clmsgid)
                    text = "joined"

                    ba = generateWebchatJSON(text, s, nick, "", srvMsg.Tags)

                    err = conn.WriteMessage(websocket.TextMessage, ba)
                    if err != nil {
                        if wsHcIrc.Debugmode {
                            fmt.Printf("[WSCHATDEBUG] ERROR sending JSON to client: %s\n", err.Error())
                        }
                    }
                }
            } else if "PART" == command {
                s = fmt.Sprintf("$.%s", channel)
                _, exists = joinedChannels[s]
                if exists {
                    s = fmt.Sprintf("sys%d", clmsgid)
                    text = fmt.Sprintf("left (after %d minutes)", (int)(srvMsg.ExtendedUserInfo.JoinDuration / 60))

                    ba = generateWebchatJSON(text, s, nick, "", srvMsg.Tags)

                    err = conn.WriteMessage(websocket.TextMessage, ba)
                    if err != nil {
                        if wsHcIrc.Debugmode {
                            fmt.Printf("[WSCHATDEBUG] ERROR sending JSON to client: %s\n", err.Error())
                        }
                    }
                }
            } else if "CLEARCHAT" == command {
                _, exists = joinedChannels[channel]
                if exists {
                    clientMsg["type"] = "clearchat"
                    clientMsg["nickId"] = wsHcIrc.NormalizeNick(text)
                    fmt.Printf("\nCLEARCHAT for %s\n\n", clientMsg["nickId"])
                    ba, err = json.Marshal(clientMsg)
                    if err != nil {
                        if wsHcIrc.Debugmode {
                            fmt.Printf("[WSCHATDEBUG] ERROR encoding JSON for client: %s\n", err.Error())
                        }
                    }
                    err = conn.WriteMessage(websocket.TextMessage, ba)
                    if err != nil {
                        if wsHcIrc.Debugmode {
                            fmt.Printf("[WSCHATDEBUG] ERROR sending JSON to client: %s\n", err.Error())
                        }
                    }
                }
            }
        }
    }

    hcIrc.UnregisterServerMessageHook(msgChanId)
    close(msgChan)

    if wsHcIrc.Debugmode {
        fmt.Printf("[WSCHATDEBUG] Connection handler terminated: %s\n", myId)
    }
}
