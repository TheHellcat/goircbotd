package hcirc

import (
    "fmt"
    "time"
    "hellcat/hcthreadutils"
)


/**
 *
 */
func (hcIrc *HcIrc) registerThread(name, id string) {
    var s string

    hcIrc.threadIds[name] = id
    s = fmt.Sprintf("Registered thread: %s (ID:%s)", name, id)
    hcIrc.debugPrint(s, "")
}


/**
 *
 */
func (hcIrc *HcIrc) inboundQueueRoutine() {
    var s string

    hcIrc.registerThread(hcthreadutils.GetCurrentName(), hcthreadutils.GetRoutineId())

    for hcIrc.inQueueRunning {
        s = hcIrc.WaitForServerMessage()
        hcIrc.InboundQueue <- s
    }

    close(hcIrc.InboundQueue)
    hcIrc.InboundQueue = nil
    hcIrc.debugPrint("Inbound queue routine ended", "")
}


/**
 *
 */
func (hcIrc *HcIrc) StartInboundQueue() {
    var inChan chan string

    hcIrc.debugPrint("Starting inbound queue", "")

    inChan = make(chan string, hcIrc.QueueSize)
    hcIrc.InboundQueue = inChan

    hcIrc.inQueueRunning = true
    go hcIrc.inboundQueueRoutine()
}


/**
 *
 */
func (hcIrc *HcIrc) StopInboundQueue() {
    hcIrc.debugPrint("Stopping inbound queue", "")
    hcIrc.inQueueRunning = false
}


/**
 *
 */
func (hcIrc *HcIrc) outboundQueueRoutine() {
    hcIrc.registerThread(hcthreadutils.GetCurrentName(), hcthreadutils.GetRoutineId())

    for s := range hcIrc.OutboundQueue {
        hcIrc.SendToServer(s)
        time.Sleep(time.Duration(hcIrc.FloodThrottle) * time.Second)
    }
    hcIrc.debugPrint("Outbound queue routine ended", "")
    hcIrc.outQueueRunning = false
    hcIrc.OutboundQueue = nil
}


/**
 *
 */
func (hcIrc *HcIrc) StartOutboundQueue() {
    var outChan chan string

    hcIrc.debugPrint("Starting outbound queue", "")

    outChan = make(chan string, hcIrc.QueueSize)
    hcIrc.OutboundQueue = outChan

    hcIrc.outQueueRunning = true
    go hcIrc.outboundQueueRoutine()
}


/**
 *
 */
func (hcIrc *HcIrc) StopOutboundQueue() {
    hcIrc.debugPrint("Stopping outbound queue", "")
    hcIrc.outQueueRunning = false
    close(hcIrc.OutboundQueue)
}


/**
 *
 */
func (hcIrc *HcIrc) outQuickQueueRoutine() {
    hcIrc.registerThread(hcthreadutils.GetCurrentName(), hcthreadutils.GetRoutineId())

    for s := range hcIrc.OutQuickQueue {
        hcIrc.SendToServer(s)
        time.Sleep((time.Duration(hcIrc.FloodThrottle) * time.Second) / 2)
    }
    hcIrc.debugPrint("Quick outbound queue routine ended", "")
    hcIrc.outQuickQueueRunning = false
    hcIrc.OutQuickQueue = nil
}


/**
 *
 */
func (hcIrc *HcIrc) StartOutQuickQueue() {
    var outChan chan string

    hcIrc.debugPrint("Starting outquick queue", "")

    outChan = make(chan string, hcIrc.QueueSize)
    hcIrc.OutQuickQueue = outChan

    hcIrc.outQuickQueueRunning = true
    go hcIrc.outQuickQueueRoutine()
}


/**
 *
 */
func (hcIrc *HcIrc) StopOutQuickQueue() {
    hcIrc.debugPrint("Stopping outquick queue", "")
    hcIrc.outQuickQueueRunning = false
    close(hcIrc.OutQuickQueue)
}
