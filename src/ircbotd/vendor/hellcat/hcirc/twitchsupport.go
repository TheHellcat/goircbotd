package hcirc

import (
    "strings"
    "strconv"
    "fmt"
    "sort"
    "encoding/json"
    "reflect"
    "net/http"
    "io/ioutil"
)

type TwitchBadgeType struct {
    ImageUrl    string
    Description string
    Title       string
    Version     string
}

var twitchBadges map[string]map[string]TwitchBadgeType
var twitchChannelBadges map[string]map[string]map[string]TwitchBadgeType


/**
 *
 */
func callHttp(url string) (string, error) {
    var r *http.Response
    var err error
    var s string
    var ba []byte

    r, err = http.Get(url)
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


/**
 *
 */
func (hcIrc *HcIrc) ParseTwitchTags(tags string) map[string]string {
    var tagList map[string]string
    var tag string
    var tagData []string

    tagList = make(map[string]string)
    tags = strings.Replace(tags, "@", "", -1)

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


/**
 *
 */
func (hcIrc *HcIrc) fetchTwitchBadges(url string) map[string]map[string]TwitchBadgeType {
    var jsonString string
    var jsonDecoder *json.Decoder
    var jMap interface{}
    var badgeData TwitchBadgeType
    var err error
    var badgeCount int
    var twitchBadgesList map[string]map[string]TwitchBadgeType

    hcIrc.debugPrint("[TWITCHSUPPORT] fetching user badge data from Twitch", "")
    badgeCount = 0

    twitchBadgesList = make(map[string]map[string]TwitchBadgeType)

    jsonString, _ = callHttp(url) //"http://badges.twitch.tv/v1/badges/global/display")
    jsonDecoder = json.NewDecoder(strings.NewReader(jsonString))

    err = jsonDecoder.Decode(&jMap)
    if err == nil {
        if ( reflect.TypeOf(jMap).String() == "map[string]interface {}") {
            for kRoot, vRoot := range jMap.(map[string]interface{}) {
                if ( "badge_sets" == kRoot ) {
                    if ( reflect.TypeOf(vRoot).String() == "map[string]interface {}") {
                        for kSets, vSets := range vRoot.(map[string]interface{}) {
                            if ( reflect.TypeOf(vSets).String() == "map[string]interface {}") {
                                for kSet, vSet := range vSets.(map[string]interface{}) {
                                    if ( "versions" == kSet ) {
                                        if ( reflect.TypeOf(vSet).String() == "map[string]interface {}") {
                                            t := make(map[string]TwitchBadgeType)
                                            for kVer, vVer := range vSet.(map[string]interface{}) {
                                                if ( reflect.TypeOf(vVer).String() == "map[string]interface {}") {
                                                    badgeData.ImageUrl = ""
                                                    badgeData.Description = ""
                                                    badgeData.Title = ""
                                                    for kDat, vDat := range vVer.(map[string]interface{}) {
                                                        if ( reflect.TypeOf(vDat).String() == "string") {
                                                            if ("image_url_1x" == kDat) {
                                                                badgeData.ImageUrl = vDat.(string)
                                                            }
                                                            if ("title" == kDat) {
                                                                badgeData.Title = vDat.(string)
                                                            }
                                                            if ("description" == kDat) {
                                                                badgeData.Description = vDat.(string)
                                                            }
                                                        }
                                                    }
                                                    badgeData.Version = kVer
                                                    t[kVer] = badgeData
                                                }
                                            }
                                            twitchBadgesList[kSets] = t
                                            hcIrc.debugPrint("[TWITCHSUPPORT] got data for badge:", kSets)
                                            badgeCount++
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    hcIrc.debugPrint("[TWITCHSUPPORT] done fetch badges. badges got:", strconv.Itoa(badgeCount))

    return twitchBadgesList
}

/**
 *
 */
func (hcIrc *HcIrc) fetchTwitchBadgesGlobal() {
    if ( len(twitchBadges) == 0 ) {
        hcIrc.debugPrint("[TWITCHSUPPORT] need global Twitch user badge data!", "")
        twitchBadges = hcIrc.fetchTwitchBadges("http://badges.twitch.tv/v1/badges/global/display")
    }
}

/**
 *
 */
func (hcIrc *HcIrc) fetchTwitchBadgesChannel(channelId string) {
    var url string

    if ( len(twitchChannelBadges) == 0 ) {
        hcIrc.debugPrint("[TWITCHSUPPORT] initialized channel user badge map", "")
        twitchChannelBadges = make(map[string]map[string]map[string]TwitchBadgeType)
    }

    if ( len(twitchChannelBadges[channelId]) == 0 ) {
        hcIrc.debugPrint("[TWITCHSUPPORT] need Twitch user badge data for channel", channelId)
        url = fmt.Sprintf("https://badges.twitch.tv/v1/badges/channels/%s/display", channelId)
        twitchChannelBadges[channelId] = hcIrc.fetchTwitchBadges(url)
    }

}


/**
 *
 */
func (hcIrc *HcIrc) ParseTwitchBadgesTag(badgesTag, channelId string) (badgesList map[int]TwitchBadgeType, count int) {
    var c int
    var a []string
    var b []string
    var s string
    var v string
    var n string
    var exists bool

    badgesList = make(map[int]TwitchBadgeType)
    c = 0

    hcIrc.fetchTwitchBadgesGlobal()
    hcIrc.fetchTwitchBadgesChannel(channelId)

    a = strings.Split(badgesTag, ",")
    for _, s = range a {
        b = strings.Split(s, "/")
        v = "1"
        if (len(b) > 1) {
            v = b[1]
        }
        n = b[0]

        badgesList[c], exists = twitchChannelBadges[channelId][n][v]
        if ( !exists ) {
            badgesList[c] = twitchBadges[n][v]
        }
        c++
    }

    count = c
    return badgesList, count
}


/**
 *
 */
func (hcIrc *HcIrc) ParseTwitchEmoteTag(emoteTag string) (emoteList map[int]TwitchMsgEmoteInfo, count int) {
    var msgEmotes string
    var emoteData []string
    var emoteInfo TwitchMsgEmoteInfo
    var emotes map[int]TwitchMsgEmoteInfo
    var emotePositions, emotePosition string
    var froms []int
    var i int
    var c int

    // "working" map to mold the emotes data into a structure at all
    emotes = make(map[int]TwitchMsgEmoteInfo)

    // the final return map (will be filled from the working one above)
    // that will have the emotes guaranteed sorted
    emoteList = make(map[int]TwitchMsgEmoteInfo)

    // first gather our emotes details into some somewhat structured data structures

    // got no emote data in the tag? nothing to do then, otherwise.... do!
    if ( len(emoteTag) > 0 ) {

        // split list of emotes in the message into an array and loop through it
        for _, msgEmotes = range strings.Split(emoteTag, "/") {

            // split ID and positions/offsets of the current emote
            emoteData = strings.Split(msgEmotes, ":")

            // ID: (saved in output struct)
            emoteInfo.Id, _ = strconv.Atoi(emoteData[0])
            // (list of) offsets where the emote occurs in the message
            emotePositions = emoteData[1]

            // now split the list of offsets into a per-occurrence array and loop through it
            for _, emotePosition = range strings.Split(emotePositions, ",") {

                // split the offset data into individual "from" and "to" values and save them in output struct
                emoteData = strings.Split(emotePosition, "-")
                emoteInfo.From, _ = strconv.Atoi(emoteData[0])
                emoteInfo.To, _ = strconv.Atoi(emoteData[1])

                // generate the image URL for displaying the emote
                emoteInfo.ChatUrl = fmt.Sprintf("https://static-cdn.jtvnw.net/emoticons/v1/%d/1.0", emoteInfo.Id)

                // save the offsets and emote data in seperate arrays, to be able to sort them later
                froms = append(froms, emoteInfo.From)
                emotes[emoteInfo.From] = emoteInfo
            }
        }
    }

    // order the positions of the emotes in the message string
    sort.Ints(froms)

    // now shove them into a SORTED (by ascending, successive indexes) map
    count = len(emotes)
    // we start the indexes for the return map with the highest one, and counting down.
    // we do this 'cause the other code in 99% of all cases will process them back-to-front in the
    // message string as otherwise it'd screw up the offsets supplied
    c = count
    for _, i = range froms {
        c--
        emoteList[c] = emotes[i]
    }

    return emoteList, count
}
