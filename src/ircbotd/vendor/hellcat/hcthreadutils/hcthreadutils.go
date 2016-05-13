package hcthreadutils

import (
    "strings"
    "runtime"
    "fmt"
    "time"
)

/*

    ****
    ****  The following code might not be the prettiest,
    ****  there is certainly way to do it more elegant,
    ****  but for now it does it's job rather well and works.
    ****

 */


type Threadinfo struct {
    Id     string
    Name   string
    State  string
    Parent string
}


/**
 *
 */
func GetRoutines() map[string]Threadinfo {
    var b []byte
    var a []string
    var i int
    var s string
    var lines []string
    var line string
    var rList map[string]Threadinfo
    var foundRoutineId bool
    var id, name, state, parent string
    var tInfo Threadinfo
    var linePadded string

    // setup our buffers and pointers
    b = make([]byte, 10240)
    rList = make(map[string]Threadinfo)

    // fetch stacktrace and make a neat array of text-lines from it
    i = runtime.Stack(b, true)
    s = string(b[:i])
    s = strings.Replace(s, string('\r'), "", -1)  // just in case....
    lines = strings.Split(s, string('\n'))

    // so, lets take that wall of text apart....:

    // start off with well defined defaults, to not confuse the loop or result data
    foundRoutineId = false
    id = ""
    name = ""
    state = ""
    parent = ""
    for _, line = range lines {
        // make a padded copy of line, to makr the slice lookups not crash on short lines but still keep the
        // original for split()ing
        linePadded = fmt.Sprintf("%s**********", line)
        if foundRoutineId {
            // on the last line we found the ID, so this line contains the name of the process
            a = strings.Split(line, ".")
            i = len(a) - 1
            s = a[i]
            a = strings.Split(s, "(")  // quick'n'dirty hack to easily get rid of the "(0xNNNNNNNN)" part xD
            s = a[0]
            name = s

            // reset our flag, to not run into this branch again until the next ID is found
            foundRoutineId = false
        }
        if "goroutine" == linePadded[:9] {
            // there is information on the next line that we want, so tell that one if above to fetch it for us
            // but first lets get ID and state of the routine/thread
            foundRoutineId = true
            a = strings.SplitN(line, " ", 3)
            id = a[1]
            state = a[2]
            i = len(state) - 2
            state = state[1:i]
        }
        if "created" == linePadded[:7] {
            // found the parent process
            a = strings.Split(line, ".")
            i = len(a) - 1
            s = a[i]
            parent = s

            // that's it, parent process is the last line of one routine stack,
            // so lets store our current information and start fresh for the next one
            tInfo.Parent = parent
            tInfo.Id = id
            tInfo.Name = name
            tInfo.State = state
            rList[name] = tInfo
            id = ""
            name = ""
            state = ""
            parent = ""
        }
    }

    return rList
}


/**
 *
 */
func GetRoutineId() string {
    var b []byte
    var i int
    var s string
    var lines []string
    var a []string
    var id string

    // setup our buffer
    b = make([]byte, 10240)

    // fetch stacktrace and make a neat array of text-lines from it
    i = runtime.Stack(b, false)
    s = string(b[:i])
    s = strings.Replace(s, string('\r'), "", -1)  // just in case....
    lines = strings.Split(s, string('\n'))

    // grab the ID and off we go
    a = strings.Split(lines[0], " ")
    id = a[1]

    return id
}


/**
 *
 */
func GetCurrentName() string {
    var b []byte
    var i int
    var s string
    var lines []string
    var a []string
    var name string

    // setup our buffer
    b = make([]byte, 10240)

    // fetch stacktrace and make a neat array of text-lines from it
    i = runtime.Stack(b, false)
    s = string(b[:i])
    s = strings.Replace(s, string('\r'), "", -1)  // just in case....
    lines = strings.Split(s, string('\n'))

    // grab the function name and off we go
    a = strings.Split(lines[3], ".")
    i = len(a) - 1
    s = a[i]
    a = strings.Split(s, "(")  // quick'n'dirty hack to easily get rid of the "(0xNNNNNNNN)" part xD
    s = a[0]
    name = s

    return name
}


/**
 *
 */
func WaitForRoutinesEnd(routines []string) {
    var threadList map[string]Threadinfo
    var thread string
    var r, running bool

    running = true
    for running {
        threadList = GetRoutines()
        running = false
        for _, thread = range routines {
            _, r = threadList[thread]
            running = running || r
        }

        if running {
            // throttle this a bit, we don't want the CPU to spike to 100% just
            // because we're waiting on some threads to terminate
            time.Sleep(100 * time.Millisecond)
        }
    }
}


/**
 *
 */
func WaitForRoutinesEndByCaller(callers []string) {
    var threadList map[string]Threadinfo
    var thread Threadinfo
    var running bool
    var caller string

    running = true
    for running {
        threadList = GetRoutines()
        running = false
        for _, thread = range threadList {
            for _, caller = range callers {
                if caller == thread.Parent {
                    running = true
                }
            }
        }

        if running {
            // throttle this a bit, we don't want the CPU to spike to 100% just
            // because we're waiting on some threads to terminate
            time.Sleep(100 * time.Millisecond)
        }
    }
}


/**
 *
 */
func WaitForRoutinesEndById(ids []string) {
    var threadList map[string]Threadinfo
    var thread Threadinfo
    var running bool
    var id string

    running = true
    for running {
        threadList = GetRoutines()
        running = false
        for _, thread = range threadList {
            for _, id = range ids {
                if id == thread.Id {
                    running = true
                }
            }
        }

        if running {
            // throttle this a bit, we don't want the CPU to spike to 100% just
            // because we're waiting on some threads to terminate
            time.Sleep(100 * time.Millisecond)
        }
    }
}
