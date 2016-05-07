package hcirc

import (
    "time"
)


/**
 *
 */
func (hcIrc *HcIrc) inboundQueueRoutine() {
    var s string

    for hcIrc.inQueueRunning {
        s = hcIrc.WaitForServerMessage()
        hcIrc.InboundQueue <- s
    }

    close(hcIrc.InboundQueue)
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
    for s := range hcIrc.OutboundQueue {
        hcIrc.SendToServer(s)
        time.Sleep(time.Duration(hcIrc.FloodThrottle) * time.Second)
    }
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
    for s := range hcIrc.OutQuickQueue {
        hcIrc.SendToServer(s)
        time.Sleep( (time.Duration(hcIrc.FloodThrottle) * time.Second)/2 )
    }
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
