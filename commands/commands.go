package commands

import (
	"fmt"
	"sort"
	"strings"

	"../cards"
	"../types"
	"../version"
)

// Constants
const (
	CmdChar = "%"
)

// CmdFuncType Command function type
type CmdFuncType func(types.API, *types.Message, []string)

// CmdFuncHelpType The type stored in the CmdFuncs map to map a function and helper text to a command
type CmdFuncHelpType struct {
	Function           CmdFuncType
	Help               string
	AllowedChannelOnly bool
}

// CmdFuncsType The type of the CmdFuncs map
type CmdFuncsType map[string]CmdFuncHelpType

// cmdFuncs Commands to functions map
var cmdFuncs CmdFuncsType

// Initializes the cmds map
func init() {
	cmdFuncs = CmdFuncsType{
		"help":     CmdFuncHelpType{cmdHelp, "Prints this list", false},
		"lookatme": CmdFuncHelpType{cmdHere, "Fuck off, user", false},
		"fuckoff":  CmdFuncHelpType{cmdNotHere, "Fuck off, bot", true},
		"version":  CmdFuncHelpType{cmdVersion, "Outputs the current bot version", true},
		"stats":    CmdFuncHelpType{cmdStats, "Displays stats about this bot", true},
		"card":     CmdFuncHelpType{cmdCard, "IS A CARD", false},
	}
}

func HandleCommand(api types.API, message *types.Message) {
	args := strings.Split(message.Content, " ")
	if len(args) == 0 {
		return
	}
	CmdFuncHelpPair, ok := cmdFuncs[args[0]]

	if ok {
		if !CmdFuncHelpPair.AllowedChannelOnly {
			CmdFuncHelpPair.Function(api, message, args)
		}
	} else {
		var reply = fmt.Sprintf("I do not have command `%s`", args[0])
		api.Reply(&types.Message{Content: reply, ChannelID: message.ChannelID})
	}
}

func cmdHelp(api types.API, message *types.Message, args []string) {
	// Build array of the keys in CmdFuncs
	var keys []string
	for k := range cmdFuncs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build message (sorted by keys) of the commands
	var cmds = "Command notation: \n`" + CmdChar + "[command] [arguments]`\n"
	cmds += "Commands:\n```\n"
	for _, key := range keys {
		cmds += fmt.Sprintf("%s - %s\n", key, cmdFuncs[key].Help)
	}
	cmds += "```\n"
	api.Reply(&types.Message{Content: cmds, ChannelID: message.ChannelID})
}

func cmdVersion(api types.API, message *types.Message, args []string) {
	reply := fmt.Sprintf("Version: %v", version.Version)
	api.Reply(&types.Message{Content: reply, ChannelID: message.ChannelID})
}

func cmdHere(api types.API, message *types.Message, args []string) {
	api.Reply(&types.Message{Content: "Here not implemented yet", ChannelID: message.ChannelID})
}

func cmdNotHere(api types.API, message *types.Message, args []string) {
	api.Reply(&types.Message{Content: "NotHere not implemented yet", ChannelID: message.ChannelID})
}

func cmdStats(api types.API, message *types.Message, args []string) {
	/*
		leaders, losers := rate.GetRespec()
		var stats = "Leaderboard:\n```\n"
		stats += leaders
		stats += "```"
		stats += "\nLosers:` "
		stats += strings.Join(losers, ", ")
		stats += " `"
		state.SendReply(message.ChannelID, stats)
	*/
	api.Reply(&types.Message{Content: "Stats not implemented yet", ChannelID: message.ChannelID})
}

func cmdCard(api types.API, message *types.Message, args []string) {
	card := cards.GenerateCard()
	api.Reply(&types.Message{Content: card.String(), ChannelID: message.ChannelID})
}
