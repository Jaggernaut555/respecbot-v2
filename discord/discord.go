package discord

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"../commands"
	"../logging"
	"../types"
	"github.com/bwmarrin/discordgo"
)

type discord struct {
	*discordgo.Session
}

func (d discord) String() string {
	return "discord"
}

var _ types.API = (*discord)(nil)
var session discord

func New(token string) types.API {
	if token == "" {
		logging.Log("You must provide a Discord authentication token (-t)")
		return nil
	}

	var err error
	session.Session, err = discordgo.New("Bot " + token)
	if err != nil {
		logging.Log("error creating Discord session,", err.Error())
		return nil
	}

	return &session
}

func (d *discord) Setup() error {
	logging.Log("Setting up respecbot on discord")
	// add a handler for when messages are posted
	d.Session.AddHandler(messageCreate)
	d.Session.AddHandler(reactionAdd)
	d.Session.AddHandler(reactionRemove)

	err := d.Session.Open()
	if err != nil {
		logging.Log("error opening connection,", err.Error())
		return err
	}
	return nil
}

func (d *discord) Listen() error {
	logging.Log("Discord api listening")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return d.Session.Close()
}

func (d *discord) Reply(message *types.Message) error {
	_, err := d.ChannelMessageSend(message.ChannelID, message.Content)
	return err
}

func (d *discord) FindMentions(msg *types.Message) []types.User {
	return nil
}

func (d *discord) HandleCommand(message *types.Message) error {
	commands.HandleCommand(d, message)
	return nil
}

func messageCreate(ds *discordgo.Session, message *discordgo.MessageCreate) {
	// Do not talk to self
	if message.Author.ID == session.State.User.ID || message.Author.Bot {
		return
	}
	logging.Log("Message recieved")

	if strings.HasPrefix(message.Content, commands.CmdChar) {
		msg := types.Message{Content: strings.TrimPrefix(message.Content, commands.CmdChar), ChannelID: message.ChannelID}
		session.HandleCommand(&msg)
		return
	}

	reply := types.Message{Content: "Message Recieved", ChannelID: message.ChannelID}
	session.Reply(&reply)
	/*
		// rate users on everything else they get
		channel, err := session.Channel(message.ChannelID)
		if err != nil {
			return
		} else if channel != nil && state.Channels[channel.ID] == true {
			rate.RespecMessage(message.Message)
	*/
}

func reactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	logging.Log("Reaction added")
	//rate.RespecReaction(reaction.MessageReaction, true)
}

func reactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	logging.Log("Reaction removed")
	//rate.RespecReaction(reaction.MessageReaction, false)
}
