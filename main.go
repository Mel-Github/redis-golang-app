package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var counter int

//var mutex = &sync.Mutex{}

func redisIncrement() int {

	conn, err := redis.Dial("tcp", "redis-node:6379")
	//conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to redis")
	}
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.
	defer conn.Close()

	// Retrieve Redis counter
	key2, err := redis.Int(conn.Do("GET", "mycounter"))

	// This use case happens when "mycounter" key dont exist inside redis
	if err == redis.ErrNil {
		fmt.Println("key2 does not exist")

		// Create a new redis key named mycounter and set value to 0
		_, err := conn.Do("SET", "mycounter", 0)
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating redis key")
		}
		// This use case happens when key exist but we are encountered problem retrieving from redis
	} else if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving redis key")
		// Successfully retrieved redis key
	} else {
		log.Info().Str("mycounter value ", strconv.Itoa(key2)).Msg("Reusing existing key")
	}

	// Incrementing Redis counter
	_, err = conn.Do("INCR", "mycounter")

	if err != nil {
		log.Fatal().Err(err).Msg("Error incrementing counter ")
	}

	// Retrieve Redis counter
	reply, err := redis.Int(conn.Do("GET", "mycounter"))
	//reply, err := redis.StringMap(conn.Do("GET", "mycounter"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving redis counter")
	}
	return reply

}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func main() {

	log.Info().Msg("main started")

	http.HandleFunc("/home", Page)
	http.HandleFunc("/health", Health)

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}

}

//Health check
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

//Page main page logic
func Page(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving current directory")
	}

	log.Info().Str("OS pwd Path", wd).Msg("Template path")

	t, err := template.ParseFiles(filepath.Join("static", "index.html"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving workspace")
	}

	// Ensure we are in the correct static template path
	log.Info().Str("Template Path", filepath.Join(wd, "./static/index.html")).Msg("successfully retrieved")

	tmpl := template.Must(t, nil)

	// Execute go template
	visitorcounter := redisIncrement()
	myvar := map[string]interface{}{
		"MyVar": visitorcounter,
	}
	tmpl.Execute(w, myvar)
}
