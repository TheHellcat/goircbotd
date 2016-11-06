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
func (hcIrc *HcIrc) fetchTwitchBadgesGlobal() {
    var jsonString string
    var jsonDecoder *json.Decoder
    var jMap interface{}
    var badgeData TwitchBadgeType
    var err error
    var badgeCount int

    if ( len(twitchBadges) == 0 ) {
        hcIrc.debugPrint("[TWITCHSUPPORT] fetching user badge data from Twitch", "")
        badgeCount = 0

        twitchBadges = make(map[string]map[string]TwitchBadgeType)

        jsonString, _ = callHttp("http://badges.twitch.tv/v1/badges/global/display")
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
                                                twitchBadges[kSets] = t
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
    }
}

// https://badges.twitch.tv/v1/badges/channels/<room-id>/display
/**
 *
 */
func (hcIrc *HcIrc) fetchTwitchBadgesChannel(channelId string) {
}


// https://discuss.dev.twitch.tv/t/beta-badge-api/6388
/**
 *
 */
func (hcIrc *HcIrc) ParseTwitchBadgesTag(badgesTag string) (badgesList map[int]TwitchBadgeType, count int) {
    var c int
    var a []string
    var b []string
    var s string
    var v string
    var n string

    badgesList = make(map[int]TwitchBadgeType)
    c = 0

    hcIrc.fetchTwitchBadgesGlobal()

    a = strings.Split(badgesTag, ",")
    for _, s = range a {
        b = strings.Split(s, "/")
        v = "1"
        if (len(b) > 1) {
            v = b[1]
        }
        n = b[0]

        badgesList[c] = twitchBadges[n][v]
        c++
    }

    count = c
    return badgesList, count
}


/**
 *
 */
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
        emoteData = strings.Split(emoteData[1], "-")  // TODO: check if index "1" exists
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
    c = count
    for _, i = range froms {
        c--
        emoteList[c] = emotes[i]
    }

    return emoteList, count
}
