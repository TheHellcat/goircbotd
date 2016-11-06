package hcirc

import (
    "strings"
    "strconv"
    "fmt"
    "sort"
    "encoding/json"
    "reflect"
)

type twitchBadgeType struct {
    imageUrl    string
    description string
    title       string
}

var twitchBadges map[string]map[string]twitchBadgeType


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


/**
 *
 */
func (hcIrc *HcIrc) fetchTwitchBadgesGlobal() {
    var jsonString string
    var jsonDecoder *json.Decoder
    var jMap interface{}
    var badgeData twitchBadgeType
    var err error
    var badgeCount int

    if ( len(twitchBadges) == 0 ) {
        hcIrc.debugPrint("[TWITCHSUPPORT] fetching user badge data from Twitch", "")
        badgeCount = 0

        twitchBadges = make(map[string]map[string]twitchBadgeType)

        jsonString = "{\"badge_sets\":{\"admin\":{\"versions\":{\"1\":{\"image_url_1x\":\"https://static-cdn.jtvnw.net/badges/v1/9ef7e029-4cdf-4d4d-a0d5-e2b3fb2583fe/1\",\"image_url_2x\":\"https://static-cdn.jtvnw.net/badges/v1/9ef7e029-4cdf-4d4d-a0d5-e2b3fb2583fe/2\",\"image_url_4x\":\"https://static-cdn.jtvnw.net/badges/v1/9ef7e029-4cdf-4d4d-a0d5-e2b3fb2583fe/3\",\"description\":\"TwitchAdmin\",\"title\":\"TwitchAdmin\",\"click_action\":\"none\",\"click_url\":\"\"}}},\"bits\":{\"versions\":{\"1\":{\"image_url_1x\":\"https://static-cdn.jtvnw.net/badges/v1/73b5c3fb-24f9-4a82-a852-2f475b59411c/1\",\"image_url_2x\":\"https://static-cdn.jtvnw.net/badges/v1/73b5c3fb-24f9-4a82-a852-2f475b59411c/2\",\"image_url_4x\":\"https://static-cdn.jtvnw.net/badges/v1/73b5c3fb-24f9-4a82-a852-2f475b59411c/3\",\"description\":\"\",\"title\":\"cheer1\",\"click_action\":\"visit_url\",\"click_url\":\"https://blog.twitch.tv/introducing-cheering-celebrate-together-da62af41fac6\"},\"100\":{\"image_url_1x\":\"https://static-cdn.jtvnw.net/badges/v1/09d93036-e7ce-431c-9a9e-7044297133f2/1\",\"image_url_2x\":\"https://static-cdn.jtvnw.net/badges/v1/09d93036-e7ce-431c-9a9e-7044297133f2/2\",\"image_url_4x\":\"https://static-cdn.jtvnw.net/badges/v1/09d93036-e7ce-431c-9a9e-7044297133f2/3\",\"description\":\"\",\"title\":\"cheer100\",\"click_action\":\"visit_url\",\"click_url\":\"https://blog.twitch.tv/introducing-cheering-celebrate-together-da62af41fac6\"}}}}}"
        //jsonString = "{\"badge_sets\":{\"admin\":\"test\",\"bits\":{\"versions\":{\"1\":{\"image_url_1x\":\"https://static-cdn.jtvnw.net/badges/v1/73b5c3fb-24f9-4a82-a852-2f475b59411c/1\",\"image_url_2x\":\"https://static-cdn.jtvnw.net/badges/v1/73b5c3fb-24f9-4a82-a852-2f475b59411c/2\",\"image_url_4x\":\"https://static-cdn.jtvnw.net/badges/v1/73b5c3fb-24f9-4a82-a852-2f475b59411c/3\",\"description\":\"\",\"title\":\"cheer1\",\"click_action\":\"visit_url\",\"click_url\":\"https://blog.twitch.tv/introducing-cheering-celebrate-together-da62af41fac6\"},\"100\":{\"image_url_1x\":\"https://static-cdn.jtvnw.net/badges/v1/09d93036-e7ce-431c-9a9e-7044297133f2/1\",\"image_url_2x\":\"https://static-cdn.jtvnw.net/badges/v1/09d93036-e7ce-431c-9a9e-7044297133f2/2\",\"image_url_4x\":\"https://static-cdn.jtvnw.net/badges/v1/09d93036-e7ce-431c-9a9e-7044297133f2/3\",\"description\":\"\",\"title\":\"cheer100\",\"click_action\":\"visit_url\",\"click_url\":\"https://blog.twitch.tv/introducing-cheering-celebrate-together-da62af41fac6\"}}}}}"
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
                                                t := make(map[string]twitchBadgeType)
                                                for kVer, vVer := range vSet.(map[string]interface{}) {
                                                    if ( reflect.TypeOf(vVer).String() == "map[string]interface {}") {
                                                        badgeData.imageUrl = ""
                                                        badgeData.description = ""
                                                        badgeData.title = ""
                                                        for kDat, vDat := range vVer.(map[string]interface{}) {
                                                            if ( reflect.TypeOf(vDat).String() == "string") {
                                                                if ("image_url_1x" == kDat) {
                                                                    badgeData.imageUrl = vDat.(string)
                                                                }
                                                                if ("title" == kDat) {
                                                                    badgeData.title = vDat.(string)
                                                                }
                                                                if ("description" == kDat) {
                                                                    badgeData.description = vDat.(string)
                                                                }
                                                            }
                                                        }
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


// https://discuss.dev.twitch.tv/t/beta-badge-api/6388
// sim-srvmsg @badges=broadcaster/1;color=;display-name=lazy_idler;emotes=1:14-18,25:4-8;mod=0;room-id=117085959;subscriber=0;turbo=0;user-id=117085959;user-type= :TestCat!TestCat@TestCat.tmi.twitch.tv PRIVMSG #test :123 Kappa 456 Kappa 789
/**
 *
 */
//func (hcIrc *HcIrc) ParseTwitchBadgesTag(badgesTag string) (badgesList map[string]string, count int) {
//var c int
//
//badgesList = make(map[string]string)
//
//for _, badge = range strings.Split(badgesTag, ",") {
//    badgeData = strings.Split(badge, "/")
//}
//}


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
    c = count
    for _, i = range froms {
        c--
        emoteList[c] = emotes[i]
    }

    return emoteList, count
}
