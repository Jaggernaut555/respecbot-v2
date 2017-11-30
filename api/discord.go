package api

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"../commands"
	"../db"
	"../logging"
	"../rate"
	"../types"
	"github.com/bwmarrin/discordgo"
)

type discord struct {
	*discordgo.Session
}

const discordName = "discord"

func (d discord) String() string {
	return discordName
}

var _ types.API = (*discord)(nil)
var session discord

func NewDiscord(token string) (types.API, error) {
	if token == "" {
		return nil, fmt.Errorf("You must provide a Discord authentication token (-t)")
	}

	var err error
	session.Session, err = discordgo.New("Bot " + token)
	if err != nil {
		logging.Log("error creating Discord session,", err.Error())
		return nil, err
	}

	return &session, nil
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

func (d *discord) ReplyTo(reply string, message *types.Message) error {
	_, err := d.ChannelMessageSend(message.Channel.ID, reply)
	return err
}

func (d *discord) HandleCommand(message *types.Message) error {
	commands.HandleCommand(d, message)
	return nil
}

func (d *discord) FindMentions(msg *types.Message) []types.User {
	return nil
}

func (d *discord) GetUser(UserID string) *types.User {
	return nil
}

func (d *discord) GetChannel(ChannelID string) *types.Channel {
	return nil
}

func (d *discord) GetServer(ServerID string) *types.Server {
	return nil
}

func messageCreate(ds *discordgo.Session, message *discordgo.MessageCreate) {
	// Do not talk to self
	if message.Author.ID == session.State.User.ID || message.Author.Bot {
		return
	}

	msg := createMessage(message.Message)

	if strings.HasPrefix(message.Content, commands.CmdChar) {
		msg.Content = strings.TrimPrefix(msg.Content, commands.CmdChar)
		session.HandleCommand(msg)
		return
	}

	// rate users on everything else they get
	if msg.Channel.Active {
		rate.RespecMessage(msg)
		db.NewMessage(msg)
	}
}

func reactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	logging.Log("Reaction added")
	//rate.RespecReaction(reaction.MessageReaction, true)
}

func reactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	logging.Log("Reaction removed")
	//rate.RespecReaction(reaction.MessageReaction, false)
}

func createMessage(message *discordgo.Message) *types.Message {
	msg := new(types.Message)

	author := getUser(message.Author)
	msg.Author = author
	msg.UserKey = author.Key

	channel := getChannel(message.ChannelID)
	msg.Channel = channel
	msg.ChannelKey = channel.Key

	msg.Content, _ = message.ContentWithMoreMentionsReplaced(session.Session)
	msg.Time, _ = message.Timestamp.Parse()
	msg.ID = message.ID

	msg.APIID = discordName

	return msg
}

func getUser(discordUser *discordgo.User) *types.User {
	user := db.GetUser(discordUser.ID, discordName)
	if user == nil {
		user = new(types.User)
		user.ID = discordUser.ID
		user.Name = discordUser.Username
		user.APIID = discordName
		db.NewUser(user)
	}
	return user
}

func getChannel(channelID string) *types.Channel {
	channel := db.GetChannel(channelID, discordName)
	if channel == nil {
		c, err := session.Channel(channelID)
		if err != nil {
			return nil
		}
		channel = new(types.Channel)
		channel.ID = channelID
		channel.Server = getServer(c.GuildID)
		channel.ServerKey = channel.Server.Key
		channel.APIID = discordName
		channel.Active = false
		db.NewChannel(channel)
	}
	return channel
}

func getServer(guildID string) *types.Server {
	server := db.GetServer(guildID, discordName)
	if server == nil {
		server = new(types.Server)
		server.ID = guildID
		server.APIID = discordName
		db.NewServer(server)
	}
	return server
}
