package main

import (
	"flag"
	"os"

	"github.com/Jaggernaut555/respecbot-v2/api"
	"github.com/Jaggernaut555/respecbot-v2/db"
	"github.com/Jaggernaut555/respecbot-v2/logging"
	"github.com/Jaggernaut555/respecbot-v2/rate"
	"github.com/Jaggernaut555/respecbot-v2/types"
)

// Global vars
var (
	token   string
	apiName string
	dbName  string
)

var (
	apiInstance types.API
)

func init() {
	flag.StringVar(&apiName, "api", "", "description")
	flag.StringVar(&token, "t", "", "Authentication token")
	flag.StringVar(&dbName, "db", "respecbot-v2.db", "Name of the database file to be used")
	purge := flag.Bool("purge", false, "Use this flag to remove all user data associated with this program")

	flag.Parse()

	if *purge {
		err := db.Purge()
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	logging.Log("TIME TO RESPEC")

	err := db.Setup(dbName)
	if err != nil {
		logging.Err(err)
		os.Exit(1)
	}

	rate.InitRatings()
}

func main() {
	var err error

	apiInstance, err = selectAPI()
	if err != nil {
		logging.Log("No API Selected")
		logging.Err(err)
		os.Exit(1)
	}

	logging.Log("Setting up API")
	err = apiInstance.Setup()
	if err != nil {
		logging.Log("API could not set up")
		logging.Err(err)
	}
	err = apiInstance.Listen()
	if err != nil {
		logging.Err(err)
	}
}

func selectAPI() (types.API, error) {
	switch apiName {
	case "discord":
		return api.NewDiscord(token)
	default:
		logging.Log("No valid api specified")
		return nil, nil
	}
}
