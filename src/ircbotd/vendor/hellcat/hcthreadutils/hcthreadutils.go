package hcthreadutils

import (
    "fmt"
    "strings"
    "runtime"
)


type threadinfo struct {
    id int
    name string
    state string
    parent string
}


/**
 *
 */
func getRoutines() map[string]threadinfo {
    var b []byte
    var i int
    var s string
    var lines []string
    var line []string
    var rList map[string]threadinfo
    var foundRoutineId bool
    var id, name, state, parent string

    // setup our buffers and pointers
    b = make([]byte,10240)
    rList = make(map[string]threadinfo)

    // fetch stacktrace and make a neat string from it
    i = runtime.Stack(b,true)
    s = string(b[:i])
    s = strings.Replace(s, string('\r'), "", -1)  // just in case....
    lines = strings.Split(s, string('\n'))

    // start off with well defined defaults, to not confuse the loop or result data
    foundRoutineId = false
    id = ""
    name = ""
    state = ""
    parent = ""
    for line = range lines {
    }

    return rList
}
