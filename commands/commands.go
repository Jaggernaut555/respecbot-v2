package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Jaggernaut555/respecbot-v2/cards"
	"github.com/Jaggernaut555/respecbot-v2/db"
	"github.com/Jaggernaut555/respecbot-v2/rate"
	"github.com/Jaggernaut555/respecbot-v2/scripting"
	"github.com/Jaggernaut555/respecbot-v2/types"
	"github.com/Jaggernaut555/respecbot-v2/version"
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
	Overwriteable      bool
}

// CmdFuncsType The type of the CmdFuncs map
type CmdFuncsType map[string]CmdFuncHelpType

// cmdFuncs Commands to functions map
var cmdFuncs CmdFuncsType

// Initializes the cmds map
func init() {
	cmdFuncs = CmdFuncsType{
		"help":     CmdFuncHelpType{cmdHelp, "Prints this list", false, false},
		"lookatme": CmdFuncHelpType{cmdHere, "Fuck off, user", false, false},
		"fuckoff":  CmdFuncHelpType{cmdNotHere, "Fuck off, bot", true, false},
		"version":  CmdFuncHelpType{cmdVersion, "Outputs the current bot version", true, false},
		"stats":    CmdFuncHelpType{cmdStats, "Displays leaderbaord, optionally use 'stats server' or 'stats global'", true, false},
		"card":     CmdFuncHelpType{cmdCard, "IS A CARD", true, false},
		"lua":      CmdFuncHelpType{cmdLua, "Lua", true, false},
	}
}

func AddCommand(funcName string, f CmdFuncHelpType) error {
	if f.Overwriteable == false {
		return fmt.Errorf("Added function must be Overwriteable")
	}
	if v, ok := cmdFuncs[funcName]; (ok && v.Overwriteable) || !ok {
		cmdFuncs[funcName] = f
		return nil
	}
	return fmt.Errorf("Cannot overwrite function '%v'", funcName)
}

func HandleCommand(api types.API, message *types.Message) {
	args := strings.Fields(message.Content)
	if len(args) == 0 {
		return
	}
	cmd := args[0]
	args = args[1:]
	CmdFuncHelpPair, ok := cmdFuncs[cmd]

	if ok {
		if !CmdFuncHelpPair.AllowedChannelOnly || message.Channel.Active {
			CmdFuncHelpPair.Function(api, message, args)
		}
	} else {
		var reply = fmt.Sprintf("I do not have command `%s`", cmd)
		api.ReplyTo(reply, message)
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
	api.ReplyTo(cmds, message)
}

func cmdVersion(api types.API, message *types.Message, args []string) {
	reply := fmt.Sprintf("Version: %v", version.Version)
	api.ReplyTo(reply, message)
}

func cmdHere(api types.API, message *types.Message, args []string) {
	if message.Channel.Active == true {
		api.ReplyTo("Yeah", message)
		return
	}
	message.Channel.Active = true
	db.UpdateChannel(message.Channel)
	api.ReplyTo("Fuck on me", message)
}

func cmdNotHere(api types.API, message *types.Message, args []string) {
	if message.Channel.Active == false {
		return
	}
	message.Channel.Active = false
	db.UpdateChannel(message.Channel)
}

func cmdStats(api types.API, message *types.Message, args []string) {
	var leaders string
	var losers []string
	if len(args) < 1 {
		leaders, losers = rate.GetRespec(message.Channel, types.Local)
	} else {
		switch strings.ToLower(args[0]) {
		case "global":
			leaders, losers = rate.GetRespec(message.Channel, types.Global)
		case "server":
			leaders, losers = rate.GetRespec(message.Channel, types.Guild)
		default:
			leaders, losers = rate.GetRespec(message.Channel, types.Local)
		}
	}
	var stats = "Leaderboard:\n```\n"
	stats += leaders
	stats += "```"
	stats += "\nLosers:` "
	stats += strings.Join(losers, ", ")
	stats += " `"
	api.ReplyTo(stats, message)
}

func cmdCard(api types.API, message *types.Message, args []string) {
	card := cards.GenerateCard()
	api.ReplyTo(card.String(), message)
}

func cmdLua(api types.API, message *types.Message, args []string) {
	scripting.Lua(api, message, args)
}
