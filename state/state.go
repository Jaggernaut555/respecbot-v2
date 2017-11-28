package state

import (
	"github.com/Jaggernaut555/respecbot/db"
	"github.com/bwmarrin/discordgo"
)

var (
	Session  *discordgo.Session
	Channels map[string]bool
	Servers  map[string]string
)

func init() {
	Channels = map[string]bool{}
	Servers = map[string]string{}
}

// InitChannels Get the active channels from db
func InitChannels() {
	db.LoadActiveChannels(&Channels, &Servers)
}

//SendReply Send a reply to the discord session
func SendReply(channelID string, reply string) {
	Session.ChannelMessageSend(channelID, reply)
}

//SendEmbed Send an embed to the discord session
func SendEmbed(channelID string, embed *discordgo.MessageEmbed) (msg *discordgo.Message) {
	msg, _ = Session.ChannelMessageSendEmbed(channelID, embed)
	return
}

//IsValidChannel Check if a channel is open to posting
func IsValidChannel(channelID string) bool {
	return Channels[channelID]
}

// GetGuildID get a guild id of the gien channel ID, ideally without having to create a channel
func GetGuildID(channelID string) string {
	guildID, ok := Servers[channelID]
	if ok {
		return guildID
	}
	channel, err := Session.Channel(channelID)
	if err != nil {
		return ""
	}
	Servers[channelID] = channel.GuildID
	return channel.GuildID
}

// NewEmbed Returns an embed ready to be written to
func NewEmbed() *discordgo.MessageEmbed {
	embed := new(discordgo.MessageEmbed)
	embed.Footer = new(discordgo.MessageEmbedFooter)
	embed.Thumbnail = new(discordgo.MessageEmbedThumbnail)
	embed.Author = new(discordgo.MessageEmbedAuthor)
	embed.Image = new(discordgo.MessageEmbedImage)
	embed.Provider = new(discordgo.MessageEmbedProvider)
	embed.Video = new(discordgo.MessageEmbedVideo)
	embed.Type = "rich"

	return embed
}

// GetUsersOfRole Returns a slice of all users in the given guild with the roleID given
func GetUsersOfRole(guild *discordgo.Guild, roleID string) (Users []*discordgo.User) {
	members := guild.Members
	for _, v := range members {
		for _, role := range v.Roles {
			if roleID == role {
				Users = append(Users, v.User)
				break
			}
		}
	}
	return
}
