package main

import (
	"flag"
	"os"

	"./discord"
	"./logging"
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

	//db.Setup(dbName, dbUsername, dbPassword, *purge)
	//state.InitChannels()
	//rate.InitRatings()
}

func main() {
	var err error
	logging.Log("TIME TO RESPEC")
	API = selectAPI()
	if API == nil {
		logging.Log("No API Selected")
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

func selectAPI() types.API {
	switch apiName {
	case "discord":
		return discord.New(token)
	default:
		logging.Log("No valid api specified")
		return nil
	}
}
