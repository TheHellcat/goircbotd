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
)

type wchatBufMsg struct {
    channel string
    nick    string
    nickId  string
    message string
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
            wchatBuf[wchatBufCur].channel = msg.Channel
            wchatBuf[wchatBufCur].nick = msg.Nick
            wchatBuf[wchatBufCur].nickId = wsHcIrc.NormalizeNick( msg.Nick )
            wchatBuf[wchatBufCur].message = msg.Text

            wchatBufCur++
            if wchatBufCur == wchatBufSize {
                wchatBufCur = 0
            }
        } else if "CLEARCHAT" == msg.Command {
            for i=0; i<wchatBufSize; i++ {
                if wsHcIrc.NormalizeNick( msg.Text ) == wchatBuf[i].nickId {
                    wchatBuf[i].message = ""
                }
            }
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

    running = true
    clmsgid = 0
    myId = request.RemoteAddr
    clientMsg = make(map[string]string)
    joinedChannels = make(map[string]string)

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
                if wsHcIrc.Debugmode {
                    fmt.Printf("[WSCHATDEBUG][HISTORYREPLAY] Sending buffer to client for channel %s\n", a[1])
                }
                for i != wchatBufCur + 1 {
                    if a[1] == wchatBuf[i].channel && len(wchatBuf[i].message) > 0 {
                        s = fmt.Sprintf("hist%d", i)
                        clientMsg["type"] = "chatmessage"
                        clientMsg["id"] = s
                        clientMsg["cssClass"] = "chatmessage"
                        clientMsg["nick"] = wchatBuf[i].nick
                        clientMsg["nickid"] =  wchatBuf[i].nickId
                        clientMsg["text"] = wchatBuf[i].message
                        ba, err = json.Marshal(clientMsg)
                        if err != nil {
                            if wsHcIrc.Debugmode {
                                fmt.Printf("[WSCHATDEBUG][HISTORYREPLAY] ERROR encoding JSON for client: %s\n", err.Error())
                            }
                        }
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
            }
            if "PART" == a[0] {
                delete(joinedChannels, a[1])
                if wsHcIrc.Debugmode {
                    fmt.Printf("[WSCHATDEBUG] %s unsubscribed to channel %s\n", myId, a[1])
                }
            }
            if "QUIT" == a[0] {
                running = false
            }

        case srvMsg = <-msgChan:
            command = srvMsg.Command
            channel = srvMsg.Channel
            nick = srvMsg.Nick
            text = srvMsg.Text
            if "PRIVMSG" == command {
                _, exists = joinedChannels[channel]
                if exists {
                    clmsgid += 1
                    if clmsgid > 268435455 {
                        // some cheap in32 kaboom protection by intentionally rolling over
                        // this way there is a clearly defines behaviour when we come close to the limit
                        // and no, I don't wanna use an int64 - it's also not really required here.
                        clmsgid = 1
                    }
                    clientMsg["type"] = "chatmessage"
                    clientMsg["id"] = fmt.Sprintf("msg%d", clmsgid)
                    clientMsg["cssClass"] = "chatmessage"
                    clientMsg["nick"] = nick
                    clientMsg["nickid"] = wsHcIrc.NormalizeNick( nick )
                    clientMsg["text"] = text
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
            } else if "CLEARCHAT" == command {
                _, exists = joinedChannels[channel]
                if exists {
                    clientMsg["type"] = "clearchat"
                    clientMsg["nickid"] = wsHcIrc.NormalizeNick( text )
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
