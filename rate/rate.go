package rate

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"text/tabwriter"
	"time"

	"../db"
	"../logging"
	"../types"
)

const (
	correctUsageValue = 2
	reactionValue     = 2
	mentionValue      = 3
	chatLimiter       = 166
)

const (
	badChange  = iota
	noChange   = iota
	goodChange = iota
)

var (
	totalRespec int
)

func InitRatings() {
	rand.Seed(time.Now().Unix())

	ratings := db.LoadGlobalUsers()

	logging.Log(fmt.Sprintf("loaded %v ratings", len(ratings)))

	totalRespec = db.GetTotalRespec()
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
	added, change := addRespecHelp(user, channel, rating)

	if change == badChange {
	} else if change == goodChange {
	}

	logging.Log(fmt.Sprintf("%v %+d respec", user.Name, added))
	return added
}

func addRespecHelp(user *types.User, channel *types.Channel, rating int) (addedRespec, polaritySwitch int) {
	// abs(userRating) / abs(totalRespec)
	userRespec := db.LoadUserRespec(user, channel)
	added := rating

	if userRespec != 0 && totalRespec != 0 {
		temp := math.Abs(float64(userRespec)) * math.Log(1+math.Abs(float64(userRespec))) / math.Abs(float64(totalRespec)) * 0.65

		if math.Abs(float64(userRespec)) > chatLimiter {
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

	totalRespec += added

	db.AddRespec(newRespec(user, channel, userRespec+added))

	if userRespec >= 0 && userRespec+added < 0 {
		return added, badChange
	} else if userRespec < 0 && userRespec+added >= 0 {
		return added, goodChange
	}

	return added, noChange
}

// evaluate messages
func RespecMessage(message *types.Message) {
	numRespec := applyRules(message)

	logging.Log(fmt.Sprintf("%v: %v", message.Author.Name, message.Content))

	//numRespec += respecMentions(guild.ID, author, message)
	AddRespec(message.Author, message.Channel, numRespec)
}

// get all da users in list
func getRatingsLists(channel *types.Channel, scope types.Scope) (users types.PairList) {
	switch scope {
	case types.Local:
		users = db.LoadChannelUsersRespec(channel)
	case types.Guild:
		users = db.LoadServerUsersRespec(channel)
	case types.Global:
		users = db.LoadGlobalUsersRespec()
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
