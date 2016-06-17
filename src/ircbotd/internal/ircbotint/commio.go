package ircbotint

import (
    "strings"
    "fmt"
    "net/http"
    "io/ioutil"
    "hellcat/hcirc"
)

var httpUrl string


/**
 *
 */
func SetHttpUrl(url string) {
    url = strings.Trim(url, "/")
    url = fmt.Sprintf("%s/", url)
    httpUrl = url
}


/**
 *
 */
func CallHttp( params []string ) (string, error) {
    var r *http.Response
    var err error
    var s string
    var ba []byte
    var i int

    //if len(param2) > 0 {
    //    s = fmt.Sprintf("%s%s/%s", httpUrl, param1, param2)
    //} else {
    //    s = fmt.Sprintf("%s%s", httpUrl, param1)
    //}
    for i=0; i<len(params); i++ {
        if i>0 {
            s = fmt.Sprintf( "%s/%s", s, params[i] )
        } else {
            s = params[i]
        }
    }
    s = fmt.Sprintf("%s%s", httpUrl, s)

    if hcirc.Self.Debugmode {
        fmt.Printf( "[COMMIODEBUG] Calling backend URL: %s\n", s )
    }

    r, err = http.Get(s)
    if err != nil {
        return "", err
    }

    ba, err = ioutil.ReadAll(r.Body)
    r.Body.Close()

    if err != nil {
        return "", err
    }

    s = string(ba)

    return s, nil
}
