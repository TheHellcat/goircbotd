package ircbotint

import (
    "gorilla/websocket"
    "net/http"
    "fmt"
)

var wsUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: checkOrigin,
}


func checkOrigin( r *http.Request ) bool {
    if "http://live.hellcat.net" == r.Header.Get("Origin") {
        return true
    } else {
        return false
    }
}


func test(w http.ResponseWriter, r *http.Request) {
    c, err := wsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        // TODO: Handle error
        fmt.Println(err.Error())
        return
    }
    defer c.Close()
    for {
        mt, message, err := c.ReadMessage()
        if err != nil {
            // TODO: Handle error
            break
        }
        fmt.Printf("** TEST ** ws-recv: %s\n", message)
        err = c.WriteMessage(mt, message)
        if err != nil {
            // TODO: Handle error
            break
        }
    }
}

func TestWs() {
    http.HandleFunc("/test", test)
    http.ListenAndServe("0.0.0.0:1234", nil)
}
