package manual

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Jaggernaut555/respecbot/bet"
	"github.com/Jaggernaut555/respecbot/db"
	"github.com/Jaggernaut555/respecbot/logging"
	"github.com/Jaggernaut555/respecbot/state"
	"github.com/bwmarrin/discordgo"
)

var (
	allBets  map[string]*bet.Bet
	betMuxes map[string]*sync.Mutex
	location *time.Location
)

func init() {
	allBets = make(map[string]*bet.Bet)
	betMuxes = make(map[string]*sync.Mutex)
	var err error
	location, err = time.LoadLocation("America/Vancouver")
	if err != nil {
		panic(err)
	}
}

func BetCmd(message *discordgo.Message, args []string) {
	/*
		format
		bet 50 @user1 @user2 ... (must have enough score, cap of 50?)
		one to many users in pot may accept (must have enough score)
		after at least one user has accepted, bet is active
		make sure user doesn't mention themself

		maybe just 'bet 50' and anybody can accept into the pool?
		One active bet per channel
	*/

	if len(args) < 2 || args[1] == "help" {
		reply := "```"
		reply += "'bet help' - display this message\n"
		reply += "'bet status' - display the status of an active bet\n"
		reply += "'bet [value] [@user/role/everyone] - create a bet\n"
		reply += "(No target is the same as @everyone)\n"
		reply += "'bet call' - Call the active bet\n"
		reply += "'bet drop' - Drop out of a bet\n"
		reply += "'bet lose' - Lose the bet\n"
		reply += "'bet start' - Start a bet early, otherwise it will start 2 minutes after it's made or when every target in the bet is ready\n"
		reply += "'bet cancel' - Cancel the active bet\n"
		reply += "(Only the bet creator can start/cancel the bet)"
		reply += "```"
		state.SendReply(message.ChannelID, reply)
		return
	}

	mux, ok := betMuxes[message.ChannelID]

	if !ok {
		mux = new(sync.Mutex)
		betMuxes[message.ChannelID] = mux
	}

	mux.Lock()

	if b, ok := allBets[message.ChannelID]; ok {
		activeBetCommand(mux, b, message, args[1])
	} else {
		createBet(mux, message, args)
	}

	mux.Unlock()
}

func activeBetCommand(mux *sync.Mutex, b *bet.Bet, message *discordgo.Message, cmd string) {
	// bet exists, check if user is active or able to join

	author := message.Author
	userStatus, ok := b.UserStatus[author.ID]

	if !ok {
		ok = b.Open
	}

	switch bet.GetEvent(cmd) {
	// begin bet with current active users
	case bet.Start:
		if author.ID == b.AuthorID && !b.Started {
			b.State <- bet.Message{User: author, Arg: bet.Start}
		}

	// cannot lose if not active
	case bet.Lose:
		if userStatus == bet.Playing && ok && b.Started {
			b.State <- bet.Message{User: author, Arg: bet.Lose}
		}

	// drop a bet before it starts
	case bet.Drop:
		if userStatus == bet.Playing && ok {
			if b.Started {
				b.State <- bet.Message{User: author, Arg: bet.Lose}
			} else {
				b.State <- bet.Message{User: author, Arg: bet.Drop}
			}
		}

	// validate user can call
	case bet.Call:
		if userStatus == bet.Lost && ok && !b.Started {
			b.State <- bet.Message{User: author, Arg: bet.Call}
		}

	// cannot cancel started bet
	case bet.Cancel:
		if author.ID == b.AuthorID {
			b.State <- bet.Message{User: author, Arg: bet.Cancel}
		}

	case bet.None:
		b.State <- bet.Message{User: author, Arg: bet.None}

	default:
		reply := fmt.Sprintf("Not a valid for active bet, use call/lose/start/cancel/status")
		state.SendReply(message.ChannelID, reply)
		b.State <- bet.Message{User: author, Arg: bet.None}
	}
}

func createBet(mux *sync.Mutex, message *discordgo.Message, args []string) {
	// bet does not exist, check if valid bet then create it
	// validate user has enough respec to create bet
	author := message.Author
	available := db.GetUserRespec(author)
	num, err := strconv.Atoi(args[1])
	if err != nil || num < 1 || available < num {
		reply := fmt.Sprintf("Invalid wager")
		state.SendReply(message.ChannelID, reply)
		return
	}

	b := bet.New(message)

	if b.Open || len(args) == 2 ||
		(len(message.Mentions) == 0 && len(message.MentionRoles) == 0) {
		b.Open = true
	} else {
		// check if role mentioned
		bet.AppendRoles(message, b)

		for _, v := range message.Mentions {
			if bet.UserCanBet(v, num) {
				b.UserStatus[v.ID] = bet.Lost
				b.Users[v.ID] = v
			}
		}
	}

	if len(b.Users) < 1 && !b.Open {
		reply := "No users can participate in this bet"
		state.SendReply(b.ChannelID, reply)
		return
	}

	if mux != betMuxes[message.ChannelID] {
		return
	}

	allBets[message.ChannelID] = b

	go engage(b.State, b, mux)
	go startBetTimer(b.State)

	b.State <- bet.Message{User: author, Arg: bet.Raise, Bet: num}

	reply := fmt.Sprintf("%v started a bet of %v", author.String(), num)
	logging.Log(reply)
}

// Engage goroutine to run an active bet
// this handles all the winnin' 'n stuff
func engage(c chan Message, b *Bet, mux *sync.Mutex) {
Loop:
	for i := range c {
		mux.Lock()
		switch i.Arg {
		case Raise:
			b.raise(i)
		case Call:
			b.call(i)
		case Lose:
			b.lose(i.User)
		case Drop:
			b.dropOut(i.User)
		case Start:
			b.start()
		case Cancel:
			b.cancel()
		case End:
			b.Ended = true
		case Win:
			b.betWon()
		default:
		}

		if !b.Started && !b.Open && !b.AgainstHouse && !b.Ended {
			if b.checkBetReady() {
				b.start()
			}
		} else if b.Started && !b.AgainstHouse && !b.Ended {
			b.Ended = b.checkWinner()
		} else if b.Started && b.AgainstHouse && !b.Ended {
		}

		if b.Ended || b.Cancelled {
			break Loop
		} else {
			activeBetEmbed(b)
		}
		mux.Unlock()
	}

	if b.Started && b.Ended && !b.Cancelled {
		b.betWon()
		b.recordBet()
	} else {
		b.cancel()
		b.deleteEmbed()
	}

	mux.Unlock()
}

func startBetTimer(c chan Message) {
	timer := time.NewTicker(time.Minute * 2)
	<-timer.C
	c <- Message{User: nil, Arg: Start}
}

func activeManualBetEmbed(b *Bet) {
	embed := state.NewEmbed()
	var title string

	if b.Started {
		title = fmt.Sprintf("Bet Started")
		embed.Footer.Text = fmt.Sprintf("Bet ends at %v", b.EndTime.Format("15:04:05"))

	} else {
		title = fmt.Sprintf("Bet Not Started")
		if b.Open {
			title += " (ANYONE CAN JOIN)"
		}
		embed.Footer.Text = fmt.Sprintf("Bet starts at %v", b.Time.Add(time.Minute*2).Format("15:04:05"))
	}

	embed.Title = title
	embed.Description = fmt.Sprintf("Total Pot: %v", b.TotalRespec)
	embed.URL = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	embed.Thumbnail.URL = "https://i.imgur.com/aUeMzFC.png"

	for k, v := range b.Users {
		field := new(discordgo.MessageEmbedField)
		field.Inline = true
		field.Name = v.Username
		if b.UserStatus[k] == Playing {
			field.Value = fmt.Sprintf("In (%v)", b.UserBet[k])
		} else {
			field.Value = "out"
		}
		embed.Fields = append(embed.Fields, field)
	}

	msg := state.SendEmbed(b.ChannelID, embed)

	if b.Annoucement != nil {
		b.deleteEmbed()
	}

	b.Annoucement = msg
}
