package types

import "time"

type Respec struct {
	Key        uint `gorm:"primary_key"`
	Respec     int
	User       *User `gorm:"ForeignKey:UserKey;save_associations:false"`
	UserKey    uint
	Channel    *Channel `gorm:"ForeignKey:ChannelKey;save_associations:false"`
	ChannelKey uint
	UpdatedAt  time.Time
}

type User struct {
	Key   uint `gorm:"primary_key;AUTO_INCREMENT"`
	ID    string
	Name  string
	APIID string
	Bot   bool `gorm:"-"` // This doesn't need to be stored in the database
}

type Message struct {
	Key        uint `gorm:"primary_key"`
	ID         string
	Author     *User `gorm:"ForeignKey:UserKey;save_associations:false"`
	UserKey    uint
	Content    string
	Channel    *Channel `gorm:"ForeignKey:ChannelKey;save_associations:false"`
	ChannelKey uint
	Mentions   []*User `gorm:"-"` // This doesn't need to be stored in the database
	Time       time.Time
	APIID      string
}

type Channel struct {
	Key       uint `gorm:"primary_key;AUTO_INCREMENT"`
	ID        string
	Server    *Server `gorm:"ForeignKey:ServerKey;save_associations:false"`
	ServerKey uint
	Active    bool
	LogActive bool
	APIID     string
}

type Server struct {
	Key   uint `gorm:"primary_key;AUTO_INCREMENT"`
	ID    string
	APIID string
}

type API interface {
	String() string
	Setup() error
	Listen() error
	ReplyTo(string, *Message) error
	ReplyToChannel(string, *Channel) error
	HandleCommand(*Message) error
	GetUser(string) *User
	GetChannel(string) *Channel
	GetLoggingChannels(*Server) *[]Channel
	GetServer(string) *Server
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type RespecList []*Respec

func (p RespecList) Len() int           { return len(p) }
func (p RespecList) Less(i, j int) bool { return p[i].Respec < p[j].Respec }
func (p RespecList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Scope uint

const (
	Local  Scope = iota
	Guild  Scope = iota
	Global Scope = iota
)

func (user *User) UserIn(users []*User) bool {
	for _, v := range users {
		if v.Key == user.Key {
			return true
		}
	}
	return false
}
