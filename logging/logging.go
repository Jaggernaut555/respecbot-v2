package logging

import (
	"github.com/Jaggernaut555/respecbot-v2/types"
	"log"
	"os"
	"strings"
	"time"
)

var (
	logger    *log.Logger
	errLogger *log.Logger

	// An API instance to send log messages to
	apiInstance types.API
)

func init() {
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errLogger = log.New(os.Stderr, "ERR - ", log.Ldate|log.Ltime)
}

func SetAPIInstance(instance types.API) {
	apiInstance = instance
}

//Log Log all the given data
func Log(data ...string) {
	logger.Print(data)
}

//Log that logs given data and sends to specific server channels
func LogToServer(server *types.Server, data ...string) {
	Log(data...)

	if apiInstance != nil {

		channels := apiInstance.GetLoggingChannels(server)

		for _, channel := range *channels {
			//TODO: Should deal with the data string differently
			t := time.Now()

			s :=  "```\n"
			s += t.Format("2006-01-02 15:04:05 - ")
			s += strings.Replace(strings.Join(data[:], ","), "`", "", -1) + "\n```"
			apiInstance.ReplyToChannel(s, &channel)
		}
	}
}

//Err Log given data to stderr
func Err(data ...error) {
	errLogger.Print(data)
}
