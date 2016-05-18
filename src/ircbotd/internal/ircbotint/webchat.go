package ircbotint

import (
    "gorilla/websocket"
    "net/http"
    "fmt"
    "hellcat/hcirc"
    "time"
)

var wsUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: checkOrigin,
}


var wsHcIrc *hcirc.HcIrc


/**
 * check the origin of the request if we want to allow the connection or not
 */
func checkOrigin( r *http.Request ) bool {
    if "http://live.hellcat.net" == r.Header.Get("Origin") {
        return true
    } else {
        return false
    }
}


/**
 * client receiver thrread
 *
 * waits for messages being sent from the webclient and sends them to the request handler
 */
func clientReceiver( conn *websocket.Conn, inChan chan string ) {
    var mt int
    var ba []byte
    var message string
    var err error
    var running bool

    running = true

    if wsHcIrc.Debugmode {
        fmt.Printf( "[WSDEBUG] Receiver thread started\n" )
    }

    for running {
        mt, ba, err = conn.ReadMessage()
        if err != nil {
            if wsHcIrc.Debugmode {
                fmt.Printf( "[WSDEBUG] Feiled to read from connection: %s\n", err.Error() )
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
        fmt.Printf( "[WSDEBUG] Receiver thread ended\n" )
    }
}


/**
 * main request handler
 *
 * handles the HTTP request, upgrade to WEBSOCKET and communication between bot and webclient
 */
func webchatHandler (writer http.ResponseWriter, request *http.Request) {
    var s string
    var msgChan chan hcirc.ServerMessage
    var inChan chan string
    var conn *websocket.Conn
    var err error
    var running bool
    var command, channel, nick, text string
    var myId string
    var srvMsg hcirc.ServerMessage

    running = true
    myId = request.RemoteAddr

    if wsHcIrc.Debugmode {
        fmt.Printf( "[WSDEBUG] New connection handler spawned: %s\n", myId )
    }

    // setup new channel to receive IRC server messages
    msgChan = make(chan hcirc.ServerMessage, wsHcIrc.QueueSize)
    // we need a unique ID for registering our channel
    s = fmt.Sprintf( "webchat-%s-%d", request.RemoteAddr, time.Now().Unix() )
    wsHcIrc.RegisterServerMessageHook( s, msgChan )
    if wsHcIrc.Debugmode {
        fmt.Printf( "[WSDEBUG] Registered server-messages channel with ID %s\n", s )
    }

    // set up channel for receiving messages from webchat client
    inChan = make(chan string, wsHcIrc.QueueSize)

    // upgrade the HTTP connection to WEBSOCKETS
    conn, err = wsUpgrader.Upgrade(writer, request, nil)
    if err != nil {
        if wsHcIrc.Debugmode {
            fmt.Printf( "[WSDEBUG] Upgrading HTTP to WEBSOCKETS feiled: %s\n", err.Error() )
        }
        return
    }
    defer conn.Close()

    // fork out the reader as separate routine/thread and listen on a chan
    // for it, this way we can have the read non-blocking and react on other
    // things as well while waiting for the client to send something
    go clientReceiver(conn, inChan)

    for running {
        select {
        case s = <-inChan:
            if "QUIT" == s {
                running = false
            }
        case srvMsg = <-msgChan:
            command = srvMsg.Command
            channel = srvMsg.Channel
            nick = srvMsg.Nick
            text = srvMsg.Text
            if "PRIVMSG" == command {
                s = fmt.Sprintf( "[%s] %s: %s", channel, nick, text )
                err = conn.WriteMessage(websocket.TextMessage, []byte(s))
                if err != nil {
                    // TODO: Handle error
                }
            }
        }
    }

    if wsHcIrc.Debugmode {
        fmt.Printf( "[WSDEBUG] Connection handler terminated: %s\n", myId )
    }
}


func TestWs(hcIrc *hcirc.HcIrc) {
    wsHcIrc = hcIrc
    http.HandleFunc("/webchat", webchatHandler)
    http.ListenAndServe("0.0.0.0:8088", nil)
}
