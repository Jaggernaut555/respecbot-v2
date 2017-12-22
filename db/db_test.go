package db

import (
	"testing"
	"time"

	"github.com/Jaggernaut555/respecbot-v2/types"
)

func TestDB(t *testing.T) {
	var err error
	var b bool

	err = Setup("test.db")
	if err != nil {
		t.Fatal(err)
	}

	//db.LogMode(true)

	user := new(types.User)
	user.ID = "userid"
	user.Name = "username"
	user.APIID = "test"
	NewUser(user)

	server := new(types.Server)
	server.ID = "serverid"
	server.APIID = "test"
	NewServer(server)

	channel := new(types.Channel)
	channel.ID = "chanid"
	channel.APIID = "test"
	channel.Server = server
	channel.ServerKey = server.Key
	NewChannel(channel)

	respec := new(types.Respec)
	respec.Channel = channel
	respec.ChannelKey = channel.Key
	respec.User = user
	respec.UserKey = user.Key
	respec.Respec = 300
	AddRespec(respec)

	message := new(types.Message)
	message.Author = user
	message.UserKey = user.Key
	message.Channel = channel
	message.ChannelKey = channel.Key
	message.APIID = "test"
	message.ID = "messageid"
	message.Content = "message content"
	message.Time = time.Now()
	NewMessage(message)

	users := GetServerRulingClass(server)
	if users == nil {
		t.Error("No users loaded")
	}

	top := GetServerRespecCap(server)
	if top != 131 {
		t.Errorf("Respec Cap not working. Expected %v, got %v", 131, top)
	}

	total := GetTotalRespec()
	if total != 300 {
		t.Error("GetTotal not working")
	}
	total = GetTotalServerRespec(server)
	if total != 300 {
		t.Error("GetTotalServer not working")
	}
	total = GetUserLocalRespec(user, channel)
	if total != 300 {
		t.Error("GetUserLocalRespec not working")
	}

	user2 := new(types.User)
	user2.ID = "userid2"
	user2.Name = "username2"
	user2.APIID = "test"
	NewUser(user2)
	server2 := new(types.Server)
	server2.ID = "serverid2"
	server2.APIID = "test"
	NewServer(server2)
	channel2 := new(types.Channel)
	channel2.ID = "chanid2"
	channel2.APIID = "test"
	channel2.Server = server
	channel2.ServerKey = server.Key
	NewChannel(channel2)
	respec2 := new(types.Respec)
	respec2.Channel = channel2
	respec2.User = user2
	respec2.UserKey = user2.Key
	respec2.Respec = 150
	AddRespec(respec2)

	user3 := new(types.User)
	user3.ID = "userid3"
	user3.Name = "username3"
	user3.APIID = "test"
	NewUser(user3)
	respec3 := new(types.Respec)
	respec3.Channel = channel
	respec3.User = user3
	respec3.UserKey = user3.Key
	respec3.Respec = -50
	AddRespec(respec3)

	GetLocalRespec(channel)

	GetServerRespec(server)

	GetGlobalRespec()

	GetUserServerRespec(user, server)

	GetLastRespecTime(user, channel)

	GetServerTopUser(server)

	GetServerRulingClass(server)

	GetServerLosers(server)

	GetLocalStats(channel)

	GetServerStats(server)

	GetGlobalStats()

	GetUser("userid", "test")

	GetChannel("chanid", "test")

	channel.Active = true
	UpdateChannel(channel)

	GetServer("serverid", "test")

	GetLastMessage(user, channel)

	GetUserLastMessages(user, channel, 3)

	GetChannelLastMessage(channel)

	message2 := new(types.Message)
	message3 := new(types.Message)
	*message2 = *message
	message2.Key = 0
	message2.Content = "message2 content"
	*message3 = *message
	message3.Key = 0
	message3.Content = "message3 content"
	NewMessage(message2)

	b = IsMultiPosting(message)
	if b == true {
		t.Error("User is not multi posting")
	}

	NewMessage(message3)

	b = IsMultiPosting(message)
	if b == false {
		t.Error("User is multi posting")
	}

	b = IsMessageUnique(message)
	if b == true {
		t.Error("Message should not be unique")
	}
	message.Content = "New content"
	b = IsMessageUnique(message)
	if b == false {
		t.Error("Message should be unique")
	}

	db.Close()
	err = DeleteDB("test.db")
	if err != nil {
		t.Fatal(err)
	}
}
