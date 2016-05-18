package ircbotint

import (
    "net/http"
    "hellcat/hcirc"
    "gorilla/websocket"
)


var wsUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: checkOrigin,
}


var wsHcIrc *hcirc.HcIrc


/**
 * Set up all URI handlers and start listener to accept incomming connections
 */
func EnableWebsocketServer(hcIrc *hcirc.HcIrc, binding string) {
    wsHcIrc = hcIrc
    http.HandleFunc("/webchat", webchatHandler)
    http.ListenAndServe(binding, nil)
}


/**
 * check the origin of the request if we want to allow the connection or not
 */
func checkOrigin( r *http.Request ) bool {
    // TODO: make this configurable - but in an EASY and SENSIBLE way!
    //if "http://live.hellcat.net" == r.Header.Get("Origin") {
        return true
    //} else {
    //    return false
    //}
}
