package main

import (
	"flag"
	"os"

	"./api"
	"./db"
	"./logging"
	"./rate"
	"./types"
)

// Global vars
var (
	token      string
	dbName     string
	dbPassword string
	dbUsername string
	apiName    string
)

var (
	API types.API
)

func init() {
	flag.StringVar(&apiName, "api", "", "description")
	flag.StringVar(&token, "t", "", "Authentication token")
	flag.StringVar(&dbPassword, "dbp", "", "Password for database user")
	flag.StringVar(&dbName, "db", "respecdb", "Database to use")
	flag.StringVar(&dbUsername, "dbu", "respecbot", "Username of database user")
	//purge := flag.Bool("purge", false, "Use this flag to purge the database. Must be used with -p")

	flag.Parse()

	//state.InitChannels()

	logging.Log("TIME TO RESPEC")

	err := db.Setup()
	if err != nil {
		logging.Err(err)
		os.Exit(1)
	}

	rate.InitRatings()
}

func main() {
	var err error

	API, err = selectAPI()
	if err != nil {
		logging.Log("No API Selected")
		logging.Err(err)
		os.Exit(1)
	}

	logging.Log("Setting up API")
	err = API.Setup()
	if err != nil {
		logging.Log("API could not set up")
		logging.Err(err)
	}
	err = API.Listen()
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
