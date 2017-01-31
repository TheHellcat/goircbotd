package ircbotint

import (
    "strings"
    "fmt"
    "net/http"
    "io/ioutil"
    "hellcat/hcirc"
    "time"
)

const (
    botApiRoot = "chatbotapi/"
    botApiVersion = "v1/"
)

type commApiAuth struct {
    Key    string
    Id     string
    Secret string
    Auth   string
}

var httpUrl string
var commApiAuthInfo commApiAuth
var commInitDone bool


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
func ioInit() {
    var m map[string]string
    var rs map[int]map[string]string
    var rv map[string]string
    var i int
    var s string
    var sa []string
    var r *http.Response
    var err error
    var ba []byte
    var loginOk bool
    var skipLoadMsg bool

    if commInitDone {
        return
    }

    skipLoadMsg = false

    DmCheckTable("system", "sysconf", "CREATE TABLE `sysconf` ( `key` TEXT NOT NULL UNIQUE, `value` TEXT DEFAULT '', PRIMARY KEY(key) );")

    m = make(map[string]string)
    m["key"] = "api_auth_key"
    rs, i = DmGet("system", "sysconf", []string{"value"}, m)
    if i > 0 {
        rv = rs[0]
        commApiAuthInfo.Key = rv["value"]

        m["key"] = "api_auth_id"
        rs, i = DmGet("system", "sysconf", []string{"value"}, m)
        if i > 0 {
            rv = rs[0]
            commApiAuthInfo.Id = rv["value"]
        }

        m["key"] = "api_auth_secret"
        rs, i = DmGet("system", "sysconf", []string{"value"}, m)
        if i > 0 {
            rv = rs[0]
            commApiAuthInfo.Secret = rv["value"]
        }
    } else {
        loginOk = false
        for !loginOk {
            commApiAuthInfo.Key = ioGenKey()
            s = fmt.Sprintf("%s%s%slogin/%s", httpUrl, botApiRoot, botApiVersion, commApiAuthInfo.Key)
            r, err = http.Get(s)
            if err == nil {
                ba, err = ioutil.ReadAll(r.Body)
                r.Body.Close()
                if err != nil {
                    return
                }
                s = string(ba)
                s = strings.Replace( s, "\n", "", -1 )
                s = strings.Replace( s, "\r", "", -1 )
                sa = strings.Split(s, ",")
                if "LOGIN" == sa[0] {
                    loginOk = true
                    m["key"] = "api_auth_key"
                    m["value"] = commApiAuthInfo.Key
                    DmSet("system", "sysconf", []string{"key"}, m)
                    m["key"] = "api_auth_id"
                    m["value"] = sa[1]
                    DmSet("system", "sysconf", []string{"key"}, m)
                    m["key"] = "api_auth_secret"
                    m["value"] = sa[2]
                    DmSet("system", "sysconf", []string{"key"}, m)
                    fmt.Printf("(i) successfully logged in to HTTP backend\n")
                    ioInit()
                } else {
                    i = 15
                    fmt.Printf("\n/!\\ please enable bot API login for bot key '%s'\n    retrying in %d seconds....\n", commApiAuthInfo.Key, i)
                    time.Sleep(time.Duration(i) * time.Second)
                }
            }
        }
        skipLoadMsg = true
    }

    if !skipLoadMsg {
        fmt.Printf("(i) successfully loaded HTTP backend authentication details\n")
    }
    commInitDone = true
}


/**
 *
 */
func ioGenAuth() {
    // TODO: change this to something sensible
    commApiAuthInfo.Auth = commApiAuthInfo.Secret
}

func ioGenKey() string {
    // TODO: change this to something sensible
    return "00000000000000000000000000000000"
}


/**
 *
 */
func CallHttp(params []string) (string, error) {
    var r *http.Response
    var err error
    var s string
    var ba []byte
    var i int
    var httpClient *http.Client
    var httpReq *http.Request

    ioInit()
    ioGenAuth()

    for i = 0; i < len(params); i++ {
        if i > 0 {
            s = fmt.Sprintf("%s/%s", s, params[i])
        } else {
            s = fmt.Sprintf("%s%s%s", botApiRoot, botApiVersion, params[i])
        }
    }
    s = fmt.Sprintf("%s%s", httpUrl, s)

    if nil != hcirc.Self {
        if hcirc.Self.Debugmode {
            fmt.Printf("[COMMIODEBUG] Calling backend URL: %s\n", s)
        }
    }

    httpClient = &http.Client{}
    httpReq, err = http.NewRequest("GET", s, nil)
    httpReq.Header.Add("access-key", commApiAuthInfo.Key)
    httpReq.Header.Add("access-id", commApiAuthInfo.Id)
    httpReq.Header.Add("access-auth", commApiAuthInfo.Auth)
    r, err = httpClient.Do(httpReq)
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
