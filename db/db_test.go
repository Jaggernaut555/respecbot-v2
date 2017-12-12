package db

import (
	"testing"

	"github.com/Jaggernaut555/respecbot-v2/types"
)

func TestDB(t *testing.T) {
	var err error

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
	respec.Respec = 440
	AddRespec(respec)

	users := GetServerRulingClass(server)
	if users == nil {
		t.Fatalf("No users loaded")
	}

	top := GetServerRespecCap(server)
	if top != 110 {
		t.Error("Respec Cap not working")
	}

	db.Close()
	err = DeleteDB("test.db")
	if err != nil {
		t.Fatal(err)
	}
}
