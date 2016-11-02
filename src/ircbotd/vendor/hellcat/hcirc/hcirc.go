package hcirc

import (
    "fmt"
    "net"
    "bufio"
    "strings"
    "hellcat/hcthreadutils"
    "strconv"
    "sort"
)

type userlist map[string]string

type Userinfo struct {
    NickDislpayname    string
    NickModes          string
    NickNormalizedName string
}

type ServerMessage struct {
    Command  string
    Channel  string
    Nick     string
    NickMode string
    User     string
    Host     string
    Text     string
    Raw      string
    Tags     string
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
    dataDir              string
    twitchMode           bool

    Error                string
}

type TwitchMsgEmoteInfo struct {
    From    int
    To      int
    Id      int
    ChatUrl string
}

var srvMsgHooks map[string]chan ServerMessage
var Self *HcIrc


/**
 * Internal init function, sets up magic internal stuff w/o which we couldn't work
 */
func init() {
    consoleRegisteredCommands = make(map[string]ConsoleCommandCallback)
    consoleRegisteredCommandInfos = make(map[string]string)
    srvMsgHooks = make(map[string]chan ServerMessage)
}


/**
 * Setup a new, usable instance of the IRC module.
 *
 * This call creates an instance of HcIrc, sets up all initial runtime values and returns the instance,
 * ready to IRC away.
 */
func New(serverHost, serverPort, serverUser, serverNick, serverPass string) (hcIrc *HcIrc) {
    Self = &HcIrc{
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
        dataDir: "./",
        twitchMode: false,
    }
    return Self
}


/**
 *
 */
func (hcIrc *HcIrc) SetDataDir(dir string) {
    hcIrc.dataDir = fmt.Sprintf("%s/", strings.Trim(dir, "/"))
}
func (hcIrc *HcIrc) GetDataDir() string {
    return hcIrc.dataDir
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
func (hcIrc *HcIrc) EnableTwitchMode() {
    hcIrc.twitchMode = true
    hcIrc.debugPrint("Enabling Twitch compatibility mode", "")
}

/**
 *
 */
func (hcIrc *HcIrc) IsTwitchModeEnabled() bool {
    return hcIrc.twitchMode
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
func (hcIrc *HcIrc) RegisterServerMessageHook(uid string, msgChannel chan ServerMessage) {
    srvMsgHooks[uid] = msgChannel
}


/**
 *
 */
func (hcIrc *HcIrc) UnregisterServerMessageHook(uid string) {
    delete(srvMsgHooks, uid)
}


/**
 *
 */
func (hcIrc *HcIrc) closeRegedServerMsgChannels() {
    var ch chan ServerMessage
    var id string

    if hcIrc.Debugmode {
        fmt.Printf("[IRCDEBUG] Closing remaining registered server message hook channels: ")
    }

    for id, ch = range srvMsgHooks {
        if hcIrc.Debugmode {
            fmt.Printf("%s ", id)
        }
        close(ch)
    }

    if hcIrc.Debugmode {
        fmt.Printf("- DONE\n")
    }

    srvMsgHooks = nil
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
    var tags string

    text = ""
    command = ""
    channel = ""
    nick = ""
    user = ""
    host = ""

    if len(message) < 2 {
        return command, channel, nick, user, host, text
    }

    if message[0:1] == "@" {
        // message contains Twitch tags, strip and separately parse them
        s1 = strings.SplitN(message, " ", 2)
        message = s1[1]
        tags = s1[0]
    }

    if message[0:1] == ":" {
        // message contains a source
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

    if "WHISPER" == command {
        // transparently convert Twitch "whispers" to IRC queries / PMs
        command = "PRIVMSG"  // that's pretty much it, lol. Why do they send it as "WHIPSER" in the first place?
        hcIrc.debugPrint("TWITCH mode:", "converted whisper to query/PM")
    }

    s = fmt.Sprintf("Parsed command '%s' with channel=%s, nick=%s, user=%s, host=%s (source=%s)", command, channel, nick, user, host, source)
    hcIrc.debugPrint(s, "")

    if hcIrc.twitchMode {
        user = fmt.Sprintf("%s\\%s", user, tags)
    }

    if hcIrc.AutohandleSysMsgs {
        hcIrc.HandleSystemMessages(command, channel, nick, user, host, text, message)
    }

    return command, channel, nick, user, host, text

}


/**
 *
 */
func (hcIrc *HcIrc) UnserializeMsgTags(tags string) map[string]string {
    var returnData map[string]string
    var tagArray []string
    var tagData []string
    var s string

    returnData = make(map[string]string)
    tags = tags[1:]

    tagArray = strings.Split(tags, ";")
    for _, s = range tagArray {
        tagData = strings.Split(s, "=")
        returnData[tagData[0]] = tagData[1]
    }

    return returnData
}


/**
 *
 */
func (hcIrc *HcIrc) HandleSystemMessages(command, channel, nick, user, host, text, raw string) {
    var s, t, u string
    var uList userlist
    var i int
    var a []string
    var b bool
    var msgChan chan ServerMessage
    var srvMsg ServerMessage
    var tags string

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
    if "NICK" == command {
        // user changed nickname, need to update all channel lists with new nick
        for s = range hcIrc.channelUsers {
            uList = hcIrc.channelUsers[s]
            u, b = uList[ hcIrc.NormalizeNick(nick) ]
            if b {
                t = fmt.Sprintf("%s%s", hcIrc.getUsermodeChars(u), hcIrc.stripUsermodeChars(text))
                hcIrc.channelUserPart(s, nick)
                hcIrc.channelUserJoin(s, t)
            }
        }
    }
    if "MODE" == command {
        hcIrc.changeUserMode(channel, nick, raw)
    }

    // send raw message to all registered receivers
    for _, msgChan = range srvMsgHooks {
        if hcIrc.twitchMode {
            // get separate user and tags for Twitch compatibility
            a = strings.Split(user, "\\")
            user = a[0]
            if len(a) > 1 {
                tags = a[1]
            } else {
                tags = ""
            }
        }
        srvMsg.Command = command
        srvMsg.Channel = channel
        srvMsg.Nick = nick
        srvMsg.NickMode = hcIrc.GetChannelUserMode(channel, nick)
        srvMsg.User = user
        srvMsg.Host = host
        srvMsg.Text = text
        srvMsg.Raw = raw
        srvMsg.Tags = tags
        msgChan <- srvMsg
    }
}


/**
 *
 */
func (hcIrc *HcIrc) SendToServer(message string) {
    var orgAutohandle bool
    var command, channel, text string

    if hcIrc.twitchMode {
        // we're having Twitch compatibility mode enabled, translate query/PM to proper "whisper"

        // first remember the current "automatically handle system messages" setting
        orgAutohandle = hcIrc.AutohandleSysMsgs

        // now prevent Parse() to trigger handling system messages
        hcIrc.AutohandleSysMsgs = false

        command, channel, _, _, _, text = hcIrc.ParseMessage(message)

        // restore original auto-handle setting
        hcIrc.AutohandleSysMsgs = orgAutohandle

        // is it a query/PM?
        if "PRIVMSG" == strings.ToUpper(command) {
            if len(channel) > 1 {
                if "#" != channel[0:1] {
                    // we got some text for a PRIVMSG and it's not a channel (not starting with "#",
                    // so this goes to a user as query/PM, let's reformat this for Twitch
                    message = fmt.Sprintf("PRIVMSG jtv :/w %s %s", channel, text)
                    hcIrc.debugPrint("TWITCH mode:", "converted query/PM to whisper")
                }
            }
        }
    }

    hcIrc.debugPrint("  to server <<<", message)
    message = strings.Replace(message, string('\n'), "", -1)
    message = strings.Replace(message, string('\r'), "", -1)
    message = fmt.Sprintf("%s\n", message)
    hcIrc.writer.WriteString(message)
    hcIrc.writer.Flush()
}


/**
 *
 */
func (hcIrc *HcIrc) ParseTwitchTags(tags string) map[string]string {
    var tagList map[string]string
    var tag string
    var tagData []string

    tagList = make(map[string]string)

    for _, tag = range strings.Split(tags, ";") {
        tagData = strings.Split(tag, "=")
        if (len(tagData) == 2) {
            tagList[tagData[0]] = tagData[1]
        } else {
            tagList[tagData[0]] = ""
        }
    }

    return tagList
}

func (hcIrc *HcIrc) ParseTwitchEmoteTag(emoteTag string) (emoteList map[int]TwitchMsgEmoteInfo, count int) {
    var emote string
    var emoteData []string
    var emoteInfo TwitchMsgEmoteInfo
    var emotes map[int]TwitchMsgEmoteInfo
    var froms []int
    var i int
    var c int

    // "working" map to mold the emotes data into a structure at all
    emotes = make(map[int]TwitchMsgEmoteInfo)

    // the final return map (will be filled from the working one above)
    // that will have the emotes guaranteed sorted
    emoteList = make(map[int]TwitchMsgEmoteInfo)

    // first gather our emotes details into some somewhat structured data structures
    for _, emote = range strings.Split(emoteTag, ",") {
        emoteData = strings.Split(emote, ":")
        emoteInfo.Id, _ = strconv.Atoi(emoteData[0])
        emoteData = strings.Split(emoteData[1], "-")
        emoteInfo.From, _ = strconv.Atoi(emoteData[0])
        emoteInfo.To, _ = strconv.Atoi(emoteData[1])
        emoteInfo.ChatUrl = fmt.Sprintf("https://static-cdn.jtvnw.net/emoticons/v1/%d/1.0", emoteInfo.Id)
        froms = append(froms, emoteInfo.From)
        emotes[emoteInfo.From] = emoteInfo
    }

    // order the positions of the emotes in the message string
    sort.Ints(froms)

    // now shove them into a SORTED (by ascending, successive indexes) map
    count = len(emotes)
    // we start the indexes for the return map with the highest one, and counting down.
    // we do this 'cause the other code in 99% of all cases will process them back-to-front in the
    // message string as otherwise it'd screw up the offsets supplied
    c=count
    for _, i = range froms {
        c--
        emoteList[c] = emotes[i]
    }

    return emoteList, count
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

    // reset the internal error variable
    hcIrc.Error = ""

    // resolve host IP and build full address string
    hcIrc.debugPrint("Looking up", hcIrc.host)
    ips, err = net.LookupIP(hcIrc.host)
    if err != nil {
        hcIrc.debugPrint("Error looking up DNS:", err.Error())
        hcIrc.Error = err.Error()
        return
    }
    ip = ips[0]
    hostIp = ip.String()
    hostPort = hcIrc.port
    hostAddr = fmt.Sprintf("%s:%s", hostIp, hostPort)

    hcIrc.debugPrint("Trying to connect to:", hostAddr)

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
    //hcIrc.WaitForServerMessage()

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

    if hcIrc.twitchMode {
        hcIrc.debugPrint("Requesting Twitch capabilities:", "commands membership tags")
        hcIrc.SendToServer("CAP REQ twitch.tv/commands")
        hcIrc.SendToServer("CAP REQ twitch.tv/membership")
        hcIrc.SendToServer("CAP REQ twitch.tv/tags")
        hcIrc.FloodThrottle = 3 // this is a more save default for Twitch, as the limit is 20 msgs per 30 secs.
    }

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

    // close all registered server message hook channels, this should also tell loops to properly terminate
    hcIrc.closeRegedServerMsgChannels()

    hcIrc.connection = nil
    hcIrc.reader = nil
    hcIrc.writer = nil
    hcIrc.channelUsers = nil

    hcIrc.Error = ""
}
