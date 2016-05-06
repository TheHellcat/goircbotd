package hcirc

import (
    "fmt"
    "net"
    "bufio"
    "strings"
    "time"
)

type HcIrc struct {
    host              string
    port              string
    user              string
    name              string
    nick              string
    pass              string

    Debugmode         bool
    AutohandleSysMsgs bool

    connection        net.Conn
    writer            *bufio.Writer
    reader            *bufio.Reader
    InboundQueue      chan string
    OutboundQueue     chan string
    inQueueRunning    bool
    outQueueRunning   bool

    QueueSize         int
    FloodThrottle     int
}

func init() {
}

func New(serverHost, serverPort, serverUser, serverNick, serverPass string) (hcIrc *HcIrc) {
    return &HcIrc{
        host: serverHost,
        port: serverPort,
        user: serverUser,
        name: "Real Name",
        nick: serverNick,
        pass: serverPass,
        Debugmode: false,
        AutohandleSysMsgs: true,
        connection: nil,
        writer: nil,
        reader: nil,
        QueueSize: 64,
        FloodThrottle: 2,
    }
}

func (hcIrc *HcIrc) debugPrint(s1, s2 string) {
    var s string

    if hcIrc.Debugmode {
        s = fmt.Sprintf("[IRCDEBUG] %s %s\n", s1, s2)
        s = strings.Replace(s, string('\n'), "", -1)
        s = strings.Replace(s, string('\r'), "", -1)
        fmt.Printf("%s\n", s)
    }
}


/**
 *
 */
func (hcIrc *HcIrc) WaitForServerMessage() string {
    var s string
    var err error

    s, err = hcIrc.reader.ReadString('\n')
    if err != nil {
        // TODO: handle error
    }

    s = strings.Replace(s, string('\n'), "", -1)
    s = strings.Replace(s, string('\r'), "", -1)
    hcIrc.debugPrint("from server >>>", s)

    return s
}


/**
 *
 */
func (hcIrc *HcIrc) ParseMessage(message string) (command, channel, nick, user, host, text string) {

    var s1 []string
    var s2 []string
    var s string
    var i int
    var source string

    text = ""
    command = ""
    channel = ""
    nick = ""
    user = ""
    host = ""

    if message[0:1] == ":" {
        s1 = strings.SplitN(message, " ", 2)
        message = s1[1]
        source = s1[0]
        source = source[1:]
    }

    s1 = strings.SplitN(message, ":", 2)
    if len(s1) == 2 {
        text = s1[1]
    }

    s1 = strings.Split(s1[0], " ")
    i = len(s1)
    if text == "" {
        text = s1[i - 1]
    }

    command = s1[0]
    channel = s1[1]

    s2 = strings.SplitN(source, "!", 2)
    if len(s2) == 2 {
        nick = s2[0]
        host = s2[1]
    } else {
        host = s2[0]
    }
    s2 = strings.SplitN(host, "@", 2)
    if len(s2) == 2 {
        user = s2[0]
        host = s2[1]
    }

    s = fmt.Sprintf("Parsed command '%s' with channel=%s, nick=%s, user=%s, host=%s (source=%s)", command, channel, nick, user, host, source)
    hcIrc.debugPrint(s, "")

    if hcIrc.AutohandleSysMsgs {
        hcIrc.HandleSystemMessages(command, channel, nick, user, host, text)
    }

    return command, channel, nick, user, host, text

}


/**
 *
 */
func (hcIrc *HcIrc) HandleSystemMessages(command, channel, nick, user, host, text string) {
    var s string

    if command == "PING" {
        s = fmt.Sprintf("PONG :%s", text)
        hcIrc.SendToServer(s)
    }
}


/**
 *
 */
func (hcIrc *HcIrc) SendToServer(message string) {
    //    var i int
    //    var err error

    hcIrc.debugPrint("  to server <<<", message)
    //    i, err =
    message = strings.Replace(message, string('\n'), "", -1)
    message = strings.Replace(message, string('\r'), "", -1)
    message = fmt.Sprintf("%s\n", message)
    hcIrc.writer.WriteString(message)
    hcIrc.writer.Flush()
}


/**
 *
 */
func (hcIrc *HcIrc) inboundQueueRoutine() {
    var s string

    for hcIrc.inQueueRunning {
        s = hcIrc.WaitForServerMessage()
        hcIrc.InboundQueue <- s
//        time.Sleep(hcIrc.FloodThrottle * time.Second)
    }

    close( hcIrc.InboundQueue )
}


/**
 *
 */
func (hcIrc *HcIrc) StartInboundQueue() {
    var inChan chan string

    inChan = make( chan string, hcIrc.QueueSize )
    hcIrc.InboundQueue = inChan

    hcIrc.inQueueRunning = true
    go hcIrc.inboundQueueRoutine()
}


/**
 *
 */
func (hcIrc *HcIrc) StopInboundQueue() {
    hcIrc.inQueueRunning = false
}


/**
 *
 */
func (hcIrc *HcIrc) outboundQueueRoutine() {
    for s := range hcIrc.OutboundQueue {
        hcIrc.SendToServer( s )
        time.Sleep(time.Duration(hcIrc.FloodThrottle) * time.Second)
    }
}


/**
 *
 */
func (hcIrc *HcIrc) StartOutboundQueue() {
    var outChan chan string

    outChan = make( chan string, hcIrc.QueueSize )
    hcIrc.OutboundQueue = outChan

    hcIrc.outQueueRunning = true
    go hcIrc.outboundQueueRoutine()
}


/**
 *
 */
func (hcIrc *HcIrc) StopOutboundQueue() {
    hcIrc.outQueueRunning = false
    close( hcIrc.OutboundQueue )
}


/**
 *
 */
func (hcIrc *HcIrc) Connect() {

    var connection net.Conn
    var err error
    var hostIp string
    var hostPort string
    var hostAddr string
    var ips []net.IP
    var ip net.IP
    var s string
    var i int
    var writer *bufio.Writer
    var reader *bufio.Reader

    // resolv host IP and build full address string
    ips, err = net.LookupIP(hcIrc.host)
    ip = ips[0]
    hostIp = ip.String()
    hostPort = hcIrc.port
    hostAddr = fmt.Sprintf("%s:%s", hostIp, hostPort)

    hcIrc.debugPrint("Connecting to:", hostAddr)

    // connect
    connection, err = net.Dial("tcp", hostAddr)
    hcIrc.connection = connection
    if err != nil {
        // TODO: error handling on failed connection / connection error
    }

    // init I/O handlers
    writer = bufio.NewWriter(connection)
    hcIrc.writer = writer
    reader = bufio.NewReader(connection)
    hcIrc.reader = reader

    // wait for first message from server
    hcIrc.WaitForServerMessage()

    // login/register with server
    if len(hcIrc.pass) > 1 {
        s = fmt.Sprintf("PASS %s", hcIrc.pass)
        hcIrc.SendToServer(s)
    }
    s = fmt.Sprintf("USER %s x x :%s", hcIrc.user, hcIrc.name)
    hcIrc.SendToServer(s)
    s = fmt.Sprintf("NICK %s", hcIrc.nick)
    hcIrc.SendToServer(s)

    // wait for server to be done with accepting our connection
    i = 1
    for i == 1 {
        s = hcIrc.WaitForServerMessage()
        // hcIrc.HandleSystemMessages( s )
        command, _, _, _, _, _ := hcIrc.ParseMessage(s)
        if command == "376" {
            i = 2
        }
    }

}
