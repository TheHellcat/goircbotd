package hcirc

import (
    "fmt"
    "net"
    "bufio"
    "strings"
    "hellcat/hcthreadutils"
)

type userlist map[string]string

type userinfo struct {
    NickDislpayname    string
    NickModes          string
    NickNormalizedName string
}

type HcIrc struct {
    host                 string
    port                 string
    user                 string
    name                 string
    nick                 string
    pass                 string

    Debugmode            bool
    AutohandleSysMsgs    bool

    QueueSize            int
    FloodThrottle        int

    InboundQueue         chan string
    OutboundQueue        chan string
    OutQuickQueue        chan string
    JoinedChannels       map[string]string

    connection           net.Conn
    writer               *bufio.Writer
    reader               *bufio.Reader
    inQueueRunning       bool
    outQueueRunning      bool
    outQuickQueueRunning bool
    channelUsers         map[string]userlist
    threadIds            map[string]string

    Error                string
}


func init() {
    consoleRegisteredCommands = make(map[string]consoleCommandCallback)
    consoleRegisteredCommandInfos = make( map[string]string )
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
        Error: "",
        inQueueRunning: false,
        outQueueRunning: false,
        channelUsers: make(map[string]userlist),
        threadIds: make(map[string]string),
        JoinedChannels: make(map[string]string),
    }
}


/**
 *
 */
func (hcIrc *HcIrc) SetRealname(name string) {
    hcIrc.name = name
}


/**
 *
 */
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
        hcIrc.debugPrint("ERROR reading from server ---", err.Error())
        hcIrc.Error = err.Error()
        return ""
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

    if len(message) < 2 {
        return command, channel, nick, user, host, text
    }

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
    if len(s1) > 1 {
        channel = s1[1]
    }

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

    // some cases where the channel name is in the text area (for whatever reason, like on some servers JOINs)
    if "" == channel {
        if len(text) > 1 {
            if "#" == text[0:1] {
                channel = text
                text = ""
            }
        }
    }

    command = strings.ToUpper(command)

    s = fmt.Sprintf("Parsed command '%s' with channel=%s, nick=%s, user=%s, host=%s (source=%s)", command, channel, nick, user, host, source)
    hcIrc.debugPrint(s, "")

    if hcIrc.AutohandleSysMsgs {
        hcIrc.HandleSystemMessages(command, channel, nick, user, host, text, message)
    }
    return command, channel, nick, user, host, text

}


/**
 *
 */
func (hcIrc *HcIrc) HandleSystemMessages(command, channel, nick, user, host, text, raw string) {
    var s string
    var i int
    var a []string

    // keepalive pings from the server
    if command == "PING" {
        s = fmt.Sprintf("PONG :%s", text)
        hcIrc.SendToServer(s)
    }

    // messages about people coming and going (NAMES lists, JOINs, PARTs, QUITs, etc.)
    if "353" == command {
        // NAMES list upon us joining a channel

        // first get the channel name from the raw message, it's the last parameter of the numeric command
        if ":" == raw[0:1] {
            i = 1
        } else {
            i = 0
        }
        a = strings.Split(raw, ":")
        s = a[i]
        a = strings.Split(strings.Trim(s, " "), " ")
        channel = a[len(a) - 1]

        // now add all users
        a = strings.Split(strings.Trim(text, " "), " ")
        for _, nick = range a {
            hcIrc.channelUserJoin(channel, nick)
        }
    }
    if "JOIN" == command {
        // a user entering channel
        hcIrc.channelUserJoin(channel, nick)
    }
    if "PART" == command {
        // a user leaving channel
        hcIrc.channelUserPart(channel, nick)
    }
    if "QUIT" == command {
        // user left the server altogether (i.e. needs to be "PARTed" from all channels)
        for s = range hcIrc.channelUsers {
            hcIrc.channelUserPart(s, nick)
        }
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
        hcIrc.Error = err.Error()
        hcIrc.debugPrint("ERROR establishing connection:", err.Error())
        return
    }

    hcIrc.debugPrint("Connection: Connected to server", "")

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
        command, _, _, _, _, _ := hcIrc.ParseMessage(s)
        if command == "376" {
            i = 2
        }
    }

    hcIrc.debugPrint("Connection: Registered with server", "")

}


/**
 *
 */
func (hcIrc *HcIrc) Shutdown() {
    if hcIrc.outQueueRunning {
        hcIrc.StopOutboundQueue()
    }

    if hcIrc.outQuickQueueRunning {
        hcIrc.StopOutQuickQueue()
    }

    if hcIrc.inQueueRunning {
        hcIrc.StopInboundQueue()
    }

    if hcIrc.connection != nil {
        hcIrc.connection.Close()
    }

    // wait for all threads/routines to have properly ended
    hcthreadutils.WaitForRoutinesEndById([]string{
        hcIrc.threadIds["inboundQueueRoutine"],
        hcIrc.threadIds["outboundQueueRoutine"],
        hcIrc.threadIds["outQuickQueueRoutine"]})

    hcIrc.connection = nil
    hcIrc.reader = nil
    hcIrc.writer = nil
    hcIrc.channelUsers = nil

    hcIrc.Error = ""
}