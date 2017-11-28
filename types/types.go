package types

type User struct {
	ID       string
	Username string
}

type Message struct {
	ID        string
	User      User
	Content   string
	ChannelID string
}

type Channel struct {
	ID       string
	ServerID string
}

type Server struct {
	ID string
}

type API interface {
	String() string
	Setup() error
	Listen() error
	Reply(*Message) error
	HandleCommand(*Message) error
	FindMentions(*Message) []User
}
