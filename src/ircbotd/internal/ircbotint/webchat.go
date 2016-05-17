package ircbotint

import (
    "gorilla/websocket"
    "net/http"
    "fmt"
    "hellcat/hcirc"
)

var wsUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: checkOrigin,
}

var msgChan chan string


func checkOrigin( r *http.Request ) bool {
    if "http://live.hellcat.net" == r.Header.Get("Origin") {
        return true
    } else {
        return false
    }
}


func test(w http.ResponseWriter, r *http.Request) {
    var s string
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
    msgChan = make(chan string)
    hcIrc.RegisterServerMessageHook( "wstest", msgChan )

    http.HandleFunc("/test", test)
    http.ListenAndServe("0.0.0.0:1234", nil)
}
