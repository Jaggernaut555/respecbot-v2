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

	db.LogMode(true)

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

	message := new(types.Message)
	message2 := new(types.Message)
	message3 := new(types.Message)
	message.Author = user
	message.UserKey = user.Key
	message.Channel = channel
	message.ChannelKey = channel.Key
	message.APIID = "test"
	message.ID = "messageid"
	message.Content = "message content"
	message.Time = time.Now()
	*message2 = *message
	message2.Content = "message2 content"
	*message3 = *message
	message3.Content = "message3 content"
	NewMessage(message)
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
