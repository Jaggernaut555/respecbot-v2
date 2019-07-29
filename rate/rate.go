package rate

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/Jaggernaut555/respecbot-v2/db"
	"github.com/Jaggernaut555/respecbot-v2/logging"
	"github.com/Jaggernaut555/respecbot-v2/types"
)

const (
	CorrectUsageValue = 2
	MentionValue      = 3
	OtherValue        = 2
)

var ()

func InitRatings() {
	rand.Seed(time.Now().Unix())

	ratings := db.GetGlobalUsers()

	logging.Log(fmt.Sprintf("loaded %v ratings", len(ratings)))
}

func newRespec(user *types.User, channel *types.Channel, rating int) *types.Respec {
	var respec types.Respec
	respec.Channel = channel
	respec.ChannelKey = channel.Key
	respec.User = user
	respec.User.Key = user.Key
	respec.Respec = rating
	return &respec
}

// AddRespec Add respec to the message, returns amount actually added
func AddRespec(user *types.User, channel *types.Channel, rating int) int {
	if user.Bot {
		return 0
	}
	added := addRespecHelp(user, channel, rating)
	logging.LogToServer(channel.Server, fmt.Sprintf("%v %+d respec", user.Name, added))
	return added
}

func addRespecHelp(user *types.User, channel *types.Channel, rating int) (addedRespec int) {
	// abs(userRating) / abs(totalRespec)
	userRespec := db.GetUserLocalRespec(user, channel)
	added := rating
	totalRespec := db.GetTotalServerRespec(channel.Server)

	if userRespec != 0 && totalRespec != 0 {
		temp := math.Abs(float64(userRespec)) * math.Log(1+math.Abs(float64(userRespec))) / float64(totalRespec) * 0.65

		if math.Abs(float64(userRespec)) > float64(db.GetServerRespecCap(channel.Server)) {
			if userRespec > 0 && added < 0 {
				temp = 0.01
			} else if userRespec < 0 && added > 0 {
				temp = 0.01
			}
		} else if temp > 0.15 {
			temp = 0.15
		} else if temp < 0.01 {
			temp = 0.01
		}
		if rand.Float64() < temp {
			added = -added
		}
	}

	db.AddRespec(newRespec(user, channel, userRespec+added))

	return added
}

// RespecMessage evaluate messages
func RespecMessage(message *types.Message) int {
	numRespec := applyRules(message)

	logging.Log(fmt.Sprintf("%v: %v", message.Author.Name, message.Content))

	respecMentions(message)

	return AddRespec(message.Author, message.Channel, numRespec)
}

func respecMentions(message *types.Message) {
	for _, v := range message.Mentions {
		if v.ID == message.Author.ID {
			logging.Log(fmt.Sprintf("%v mentioned themself in channel %v", message.Author, message.ChannelKey))
			AddRespec(message.Author, message.Channel, -MentionValue)
			continue
		}
		logging.Log(fmt.Sprintf("%v Mentioned %v in channel %v\n", message.Author.Name, v.Name, message.ChannelKey))
		RespecOther(v, message.Channel, MentionValue)
	}
}

// RespecOther Give respec by some other means, ie mentioning.
// Something that a user has no control and will only be applicable every 5 minutes
func RespecOther(user *types.User, channel *types.Channel, rating int) (added int) {
	now := time.Now()
	last := db.GetLastRespecTime(user, channel)
	if last != nil {
		timeDelta := now.Sub(*last)
		if timeDelta.Minutes() > 5 {
			return AddRespec(user, channel, rating)
		}
	} else {
		return AddRespec(user, channel, rating)
	}
	return 0
}

// get all da users in list
func getRatingsLists(channel *types.Channel, scope types.Scope) (users types.PairList) {
	switch scope {
	case types.Local:
		users = db.GetLocalStats(channel)
	case types.Guild:
		users = db.GetServerStats(channel.Server)
	case types.Global:
		users = db.GetGlobalStats()
	}
	return
}

// show 10 most RESPEC peep
func GetRespec(channel *types.Channel, scope types.Scope) (Leaderboard string, negativeUsers []string) {
	var buf bytes.Buffer
	negativeUsers = make([]string, 0)
	users := getRatingsLists(channel, scope)

	sort.Sort(sort.Reverse(users))

	var padding = 3
	w := new(tabwriter.Writer)
	w.Init(&buf, 0, 0, padding, ' ', 0)
	for k, v := range users {
		if k > 15 {
			break
		}
		if v.Value >= 0 {
			fmt.Fprintf(w, "%v\t%v\t\n", v.Key, v.Value)
		} else {
			negativeUsers = append(negativeUsers, v.Key)
		}
	}
	w.Flush()
	Leaderboard = fmt.Sprintf("%v", buf.String())
	sort.Strings(negativeUsers)
	return
}
