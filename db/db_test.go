package db

import (
	"os"
	"testing"

	"github.com/Jaggernaut555/respecbot-v2/logging"
	"github.com/Jaggernaut555/respecbot-v2/types"
	"github.com/jinzhu/gorm"
)

func TestDB(t *testing.T) {
	var err error
	os.Remove("test.db")
	db, err = gorm.Open("sqlite3", "test.db")
	if err != nil {
		t.Fatal(err)
	}
	logging.Log("SQLite file setup at", "test.db")

	createTables(db)
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
	respec.Respec = 5
	AddRespec(respec)

	GetServerRulingClass(server)

	users := GetServerUsers(server)
	if users == nil {
		t.Fatalf("No users loaded")
	}

	db.Close()
	err = os.Remove("test.db")
	if err != nil {
		t.Fatal(err)
	}
}
