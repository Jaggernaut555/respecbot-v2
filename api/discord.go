package api

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Jaggernaut555/respecbot-v2/commands"
	"github.com/Jaggernaut555/respecbot-v2/db"
	"github.com/Jaggernaut555/respecbot-v2/logging"
	"github.com/Jaggernaut555/respecbot-v2/rate"
	"github.com/Jaggernaut555/respecbot-v2/types"
	"github.com/bwmarrin/discordgo"
)

type discord struct {
	*discordgo.Session
}

const discordName = "discord"

const (
	supremeRoleName = "Supreme Ruler"
	rulingRoleName  = "Ruling Class"
	loserRoleName   = "Losers"
)

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

func (d *discord) GetUser(userID string) *types.User {
	return db.GetUser(userID, discordName)
}

func (d *discord) GetChannel(channelID string) *types.Channel {
	return db.GetChannel(channelID, discordName)
}

func (d *discord) GetServer(serverID string) *types.Server {
	return db.GetServer(serverID, discordName)
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
		updateServerStatus(msg.Channel.Server)
	}
}

func reactionAdd(s *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	logging.Log("Reaction added")
	message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
	if err != nil {
		logging.Err(err)
		return
	}
	author := getUser(message.Author)
	channel := getChannel(reaction.ChannelID)
	if channel.Active && reaction.UserID != author.ID {
		rate.RespecOther(author, channel, rate.OtherValue)
		updateServerStatus(channel.Server)
	}
}

func reactionRemove(s *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	logging.Log("Reaction removed")
	message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
	if err != nil {
		logging.Err(err)
		return
	}
	author := getUser(message.Author)
	channel := getChannel(reaction.ChannelID)
	if channel.Active && reaction.UserID != author.ID {
		rate.RespecOther(author, channel, -rate.OtherValue)
		updateServerStatus(channel.Server)
	}
}

func getRoleID(guildID, roleName string) (roleID string) {
	roles, _ := session.GuildRoles(guildID)
	var role *discordgo.Role
	for _, v := range roles {
		if v.Name == roleName {
			role = v
			break
		}
	}
	if role == nil {
		return ""
	}
	return role.ID
}

func updateServerStatus(server *types.Server) {
	top := db.GetServerTopUser(server)
	ruling := db.GetServerRulingClass(server)
	losers := db.GetServerLosers(server)
	users := db.GetServerUsers(server)

	for _, v := range users {
		if v.UserIn(losers) {
			makeUserLoser(server.ID, v.ID)
			makeUserNotTop(server.ID, v.ID)
			makeUserNotRuling(server.ID, v.ID)
			continue
		}
		makeUserNotLoser(server.ID, v.ID)
		if v.ID == top.ID {
			makeUserTop(server.ID, v.ID)
			makeUserRuling(server.ID, v.ID)
			continue
		} else {
			makeUserNotTop(server.ID, v.ID)
		}
		if v.UserIn(ruling) {
			makeUserRuling(server.ID, v.ID)
		} else {
			makeUserNotRuling(server.ID, v.ID)
		}
	}
}

func makeUserLoser(guildID, userID string) {
	roleID := getRoleID(guildID, loserRoleName)
	userAddRole(guildID, userID, roleID)
}

func makeUserNotLoser(guildID, userID string) {
	roleID := getRoleID(guildID, loserRoleName)
	userRemoveRole(guildID, userID, roleID)
}

func makeUserTop(guildID, userID string) {
	roleID := getRoleID(guildID, supremeRoleName)
	userAddRole(guildID, userID, roleID)
}

func makeUserNotTop(guildID, userID string) {
	roleID := getRoleID(guildID, supremeRoleName)
	userRemoveRole(guildID, userID, roleID)
}

func makeUserRuling(guildID, userID string) {
	roleID := getRoleID(guildID, rulingRoleName)
	userAddRole(guildID, userID, roleID)
}

func makeUserNotRuling(guildID, userID string) {
	roleID := getRoleID(guildID, rulingRoleName)
	userRemoveRole(guildID, userID, roleID)
}

func userAddRole(serverID, userID, roleID string) {
	session.GuildMemberRoleAdd(serverID, userID, roleID)
}

func userRemoveRole(serverID, userID, roleID string) {
	session.GuildMemberRoleRemove(serverID, userID, roleID)
}

func createMessage(message *discordgo.Message) *types.Message {
	msg := new(types.Message)

	author := getUser(message.Author)
	msg.Author = author
	msg.UserKey = author.Key

	channel := getChannel(message.ChannelID)
	msg.Channel = channel
	msg.ChannelKey = channel.Key

	msg.Mentions = getMentionedUsers(message, msg)

	msg.Content, _ = message.ContentWithMoreMentionsReplaced(session.Session)
	msg.Time, _ = message.Timestamp.Parse()
	msg.ID = message.ID

	msg.APIID = discordName

	return msg
}

func getMentionedUsers(message *discordgo.Message, msg *types.Message) []*types.User {
	var users []*types.User
	userMap := make(map[string]*types.User)

	for _, v := range message.Mentions {
		userMap[v.ID] = getUser(v)
	}

	for _, v := range message.MentionRoles {
		roleUsers := getMentionedRoles(msg, v)
		for _, v := range roleUsers {
			userMap[v.ID] = v
		}
	}

	for _, v := range userMap {
		users = append(users, v)
	}

	return users
}

func getMentionedRoles(msg *types.Message, roleID string) []*types.User {
	var users []*types.User
	guild, err := session.Guild(msg.Channel.Server.ID)
	if err != nil {
		logging.Err(err)
		return nil
	}
	for _, v := range guild.Members {
		for _, role := range v.Roles {
			if roleID == role {
				users = append(users, getUser(v.User))
			}
		}
	}

	return users
}

func getUser(discordUser *discordgo.User) *types.User {
	user := db.GetUser(discordUser.ID, discordName)
	if user == nil {
		user = new(types.User)
		user.ID = discordUser.ID
		user.Name = discordUser.Username
		user.APIID = discordName
		user.Bot = discordUser.Bot
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
