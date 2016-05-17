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


func checkOrigin( r *http.Request ) bool {
    if "http://live.hellcat.net" == r.Header.Get("Origin") {
        return true
    } else {
        return false
    }
}


func test(w http.ResponseWriter, r *http.Request) {
    var s string
    var msgChan chan string

    fmt.Println( "WS TEST HANDLER" )

    msgChan = make(chan string)
    s = fmt.Sprintf( "swtest-%s-%s", r.RemoteAddr, time.Now().String() )
    fmt.Println( s )
    wsHcIrc.RegisterServerMessageHook( s, msgChan )

    c, err := wsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        // TODO: Handle error
        fmt.Println(err.Error())
        return
    }
    defer c.Close()
    //for {
        mt, message, err := c.ReadMessage()
    message = message
        if err != nil {
            // TODO: Handle error
            //break
            return
        }
        fmt.Println( mt )
        for {
            s = <-msgChan
            //fmt.Printf("** TEST ** ws-recv: %s\n", s)
            err = c.WriteMessage(mt, []byte(s))
            if err != nil {
                // TODO: Handle error
                break
            }
        }
    //}
}

func TestWs(hcIrc *hcirc.HcIrc) {
    wsHcIrc = hcIrc
    http.HandleFunc("/test", test)
    http.ListenAndServe("0.0.0.0:1234", nil)
}
