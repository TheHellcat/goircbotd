package ircbotint

import (
    "strings"
    "fmt"
    "net/http"
    "io/ioutil"
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
func CallHttp(param1, param2 string) (string, error) {
    var r *http.Response
    var err error
    var s string
    var ba []byte

    if len(param2) > 0 {
        s = fmt.Sprintf("%s%s/%s", httpUrl, param1, param2)
    } else {
        s = fmt.Sprintf("%s%s", httpUrl, param1)
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
