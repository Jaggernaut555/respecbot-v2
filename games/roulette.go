package games

import (
	"sync"

	"github.com/Jaggernaut555/respecbot/bet"
	"github.com/Jaggernaut555/respecbot/state"
	"github.com/bwmarrin/discordgo"
)

type roulette struct {
	bet.Bet
}

var (
	roulettes     map[string]roulette
	rouletteMuxes map[string]*sync.Mutex
)

func init() {
	roulettes = make(map[string]roulette)
	rouletteMuxes = make(map[string]*sync.Mutex)
}

func rouletteCmd(message *discordgo.Message, args []string) {
	if len(args) < 2 || args[1] == "help" {
		reply := "```"
		reply += "under construction\n"
		reply += "```"
		state.SendReply(message.ChannelID, reply)
		return
	}
}
