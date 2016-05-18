package ircbotint

import (
    "net/http"
    "hellcat/hcirc"
    "gorilla/websocket"
    "fmt"
    "net"
)


type tcpKeepAliveListener struct {
    *net.TCPListener
}

var wsListener net.Listener

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
    fmt.Printf( "(i) starting HTTP/Websockets listener\n" )
    listenAndServe(binding, nil)
    fmt.Printf( "(i) HTTP/Websockets terminated\n" )
    wsListener = nil
}


/**
 * Shut down the http/ws listener
 */
func DisableWebsocketServer() {
    if wsListener != nil {
        wsListener.Close()
    }
}


/**
 * Register additional commands for the internal, interactive console.
 * This allows you to control certain aspects (like starting/stopping) of the listener
 * from the console
 */
func RegisterWebsocketConsoleCommands() {
    wsHcIrc.RegisterConsoleCommand( "ws-enable", "Enable http/websocket listener", consEnableWs )
    wsHcIrc.RegisterConsoleCommand( "ws-disable", "Disable http/websocket listener", consDisableWs )
}


/**
 * Function for console command to start/enable the http/ws listener
 */
func consEnableWs( cmd, param string ) string {
    var r string

    if len(param) < 10 {
        r = "You must provide a proper binding to listing on,\n e.g. 0.0.0.0:8000, 127.0.0.1:8000, 192.168.123.45:1234 or similar.\n"
    } else {
        go EnableWebsocketServer( hcirc.Self, param )
        r = fmt.Sprintf( "Enabling http/ws listener on %s\n", param )
    }

    return r
}


/**
 * Function for console command to stop/disable the http/ws listener
 */
func consDisableWs( cmd, param string ) string {
    DisableWebsocketServer()
    return "Shutting down http/ws listener\n"
}


/**
 * Internal function to actually create the net listener instance, start listening and
 * serving the requests
 *
 * It's mostly a copy-paste from the http.ListenAndServe function, modified a bit to my
 * own needs, esp. the ability to shut down the listener (the original http.ListenAndServe,
 * once running, knows no return)
 */
func listenAndServe( addr string, handler http.Handler) {
    var err error
    var srv *http.Server

    srv = &http.Server{Addr: addr, Handler: handler}
    if addr == "" {
        addr = ":http"
    }
    wsListener, err = net.Listen("tcp", addr)
    if err != nil {
        fmt.Printf( "(!) Failed to open listener: %s\n", err.Error() )
        return
    }
    err = srv.Serve(tcpKeepAliveListener{wsListener.(*net.TCPListener)})
    if err != nil {
        fmt.Printf( "(!) Error during listening: %s\n", err.Error() )
    }
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
