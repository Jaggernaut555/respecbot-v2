package db

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Jaggernaut555/respecbot-v2/logging"
	"github.com/Jaggernaut555/respecbot-v2/types"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/shibukawa/configdir"
)

const (
	dbFileName  = "respecbot-v2.db"
	projectName = "respecbot-v2"
	vendorName  = "Jaggernaut555"
)

var db *gorm.DB
var fileDir *configdir.Config
var dbFile string

func Setup() error {
	var err error

	configDir := configdir.New(vendorName, projectName)
	fileDir = configDir.QueryCacheFolder()

	dbFile = filepath.FromSlash(fileDir.Path + "/" + dbFileName)

	if err = fileDir.MkdirAll(); err != nil {
		return err
	}

	if !fileDir.Exists(dbFileName) {
		if _, err = fileDir.Create(dbFileName); err != nil {
			return err
		}
	}

	db, err = gorm.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}
	logging.Log("SQLite file setup at", dbFile)

	createTables(db)

	//db.LogMode(true)

	return err
}

func Purge() error {
	configDir := configdir.New(vendorName, projectName)
	fileDir = configDir.QueryCacheFolder()
	logging.Log(fmt.Sprintf("Deleting %v", fileDir.Path))
	return os.RemoveAll(fileDir.Path)
}

func createTables(d *gorm.DB) {
	if !d.HasTable(&types.User{}) {
		d.CreateTable(&types.User{})
	}
	if !d.HasTable(&types.Channel{}) {
		d.CreateTable(&types.Channel{})
	}
	if !d.HasTable(&types.Server{}) {
		d.CreateTable(&types.Server{})
	}
	if !d.HasTable(&types.Message{}) {
		d.CreateTable(&types.Message{})
	}
	if !d.HasTable(&types.Respec{}) {
		d.CreateTable(&types.Respec{})
	}
}

func GetTotalRespec() int {
	var total []types.Respec
	db.Model(&types.Respec{}).Select("sum(Respec) as respec").Scan(&total)
	return total[0].Respec
}

func GetTotalServerRespec(server *types.Server) int {
	var total []types.Respec
	db.Model(&types.Respec{}).Preload("Channel.Server", "key = ?", server.Key).Select("sum(Respec) as respec").Scan(&total)
	return total[0].Respec
}

func LoadGlobalUsers() []*types.User {
	var users []*types.User
	if db.Find(&users).RecordNotFound() {
		return nil
	}
	return users
}

func LoadServerRespec(server *types.Server) []*types.Respec {
	var respec []*types.Respec
	if db.Preload("User").Preload("Channel.Server", "key = ?", server.Key).Find(&respec).RecordNotFound() {
		return nil
	}
	return respec
}

func LoadGlobalRespec() []*types.Respec {
	var respec []*types.Respec
	if db.Preload("User").Preload("Channel").Preload("Channel.Server").Find(&respec).RecordNotFound() {
		return nil
	}
	return respec
}

func LoadUserRespec(user *types.User, channel *types.Channel) int {
	var respec types.Respec
	if db.First(&respec, types.Respec{UserKey: user.Key, ChannelKey: channel.Key}).RecordNotFound() {
		return 0
	}
	return respec.Respec
}

func GetLastRespecTime(user *types.User, channel *types.Channel) *time.Time {
	var respec types.Respec
	if db.First(&respec, types.Respec{UserKey: user.Key, ChannelKey: channel.Key}).RecordNotFound() {
		return nil
	}
	return &respec.UpdatedAt
}

func AddRespec(respec *types.Respec) {
	db.Where(types.Respec{UserKey: respec.User.Key, ChannelKey: respec.Channel.Key}).Assign(types.Respec{Respec: respec.Respec}).FirstOrCreate(respec)
}

func LoadChannelStats(channel *types.Channel) types.PairList {
	var pairs types.PairList
	var respec []*types.Respec
	if db.Preload("User").Find(&respec, types.Respec{ChannelKey: channel.Key}).RecordNotFound() {
		return nil
	}

	for _, v := range respec {
		pairs = append(pairs, types.Pair{Key: v.User.Name, Value: v.Respec})
	}

	return pairs
}

func LoadServerStats(channel *types.Channel) types.PairList {
	var pairs types.PairList
	var respec []*types.Respec
	if db.Table("respecs a").Preload("User").Preload("Channel", "server_key = ?", channel.Server.Key).Group("a.user_key").Select("a.user_key, sum(a.respec) as respec").Find(&respec).RecordNotFound() {
		return nil
	}

	for _, v := range respec {
		pairs = append(pairs, types.Pair{Key: v.User.Name, Value: v.Respec})
	}

	return pairs
}

func LoadGlobalStats() types.PairList {
	var pairs types.PairList
	var respec []*types.Respec
	if db.Table("respecs a").Preload("User").Group("a.user_key").Select("a.user_key, sum(a.respec) as respec").Find(&respec).RecordNotFound() {
		return nil
	}

	for _, v := range respec {
		pairs = append(pairs, types.Pair{Key: v.User.Name, Value: v.Respec})
	}

	return pairs
}

func LoadServerUsersRespecs(channel *types.Channel) types.PairList {
	var pairs types.PairList
	var respec []*types.Respec
	if db.Table("respecs a").Preload("User").Preload("Channel", "server_key = ?", channel.Server.Key).Group("a.user_key").Select("a.user_key, sum(a.respec) as respec").Find(&respec).RecordNotFound() {
		return nil
	}

	for _, v := range respec {
		pairs = append(pairs, types.Pair{Key: v.User.Name, Value: v.Respec})
	}

	return pairs
}

func NewUser(user *types.User) error {
	if user.APIID == "" {
		return fmt.Errorf("APIID not set")
	}
	if user.Name == "" {
		return fmt.Errorf("Name not set")
	}
	if user.ID == "" {
		return fmt.Errorf("ID not set")
	}
	if user.Key != 0 {
		return fmt.Errorf("Key already set")
	}
	if user.Bot {
		return fmt.Errorf("Cannot add bot user")
	}
	if db.NewRecord(user) {
		db.Create(user)
	}
	return nil
}

func GetUser(UserID, APIID string) *types.User {
	var user types.User
	if db.Where("id = ? AND api_id = ?", UserID, APIID).First(&user).RecordNotFound() {
		return nil
	}
	return &user
}

func NewChannel(channel *types.Channel) {
	if db.NewRecord(channel) {
		db.Create(channel)
	}
}

func GetChannel(channelID, APIID string) *types.Channel {
	var channel types.Channel
	if db.Preload("Server").Where("id = ? AND api_id = ?", channelID, APIID).First(&channel).RecordNotFound() {
		return nil
	}
	return &channel
}

func UpdateChannel(channel *types.Channel) {
	db.Save(channel)
}

func NewServer(server *types.Server) {
	if db.NewRecord(server) {
		db.Create(server)
	}
}

func GetServer(serverID, APIID string) *types.Server {
	var server types.Server
	if db.Where("id = ? AND api_id = ?", serverID, APIID).First(&server).RecordNotFound() {
		return nil
	}
	return &server
}

func NewMessage(message *types.Message) {
	if db.NewRecord(message) {
		db.Create(message)
	}
}

func GetLastMessage(user *types.User, channel *types.Channel) *types.Message {
	var message types.Message
	if db.Preload("Channel").Preload("Channel.Server").Preload("Author").Where("user_key = ? AND channel_key = ?", user.Key, channel.Key).Order("time desc").First(&message).RecordNotFound() {
		return nil
	}
	return &message
}

func GetLastFiveMessages(user *types.User, channel *types.Channel) []*types.Message {
	var messages []*types.Message
	if db.Preload("Channel").Preload("Channel.Server").Preload("Author").Where("user_key = ? AND channel_key = ?", user.Key, channel.Key).Order("time desc").Limit(5).Find(&messages).RecordNotFound() {
		return nil
	}
	return messages
}

func GetChannelLastMessage(channel *types.Channel) *types.Message {
	var message types.Message
	if db.Preload("Channel").Preload("Channel.Server").Preload("Author").Where("channel_key = ?", channel.Key).Order("time desc").First(&message).RecordNotFound() {
		return nil
	}
	return &message
}
