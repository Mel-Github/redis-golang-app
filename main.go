package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var counter int
var mutex = &sync.Mutex{}

func echoString(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}

func incrementCounter() int {
	mutex.Lock()

	//TODO write to redis
	counter := redisIncrement()

	log.Print("Counter stand " + strconv.Itoa(counter))
	mutex.Unlock()
	return counter
}

func redisIncrement() int {

	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to redis")
	}
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.
	defer conn.Close()

	// Incrementing Redis counter
	_, err = conn.Do("INCR", "mycounter")

	if err != nil {
		log.Fatal().Err(err).Msg("Error incrementing counter ")
	}

	// Retrieve Redis counter
	reply, err := redis.Int(conn.Do("GET", "mycounter"))
	//reply, err := redis.StringMap(conn.Do("GET", "mycounter"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving album")
	}

	log.Info().Str("Current counter value", strconv.Itoa(reply)).Msg("Current counter value successfully retrieved")

	return reply

}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func main() {

	log.Info().Msg("main started")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving workspace")
	}

	//log.Info().Str("path", wd).Msg("Template path")

	t, err := template.ParseFiles(filepath.Join(wd, "./static/index.html"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error retrieving workspace")
	}

	// Ensure we are in the correct static template path
	log.Info().Str("Template Path", filepath.Join(wd, "./static/index.html")).Msg("successfully retrieved")

	tmpl := template.Must(t, nil)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		visitorcounter := incrementCounter()
		myvar := map[string]interface{}{
			"MyVar": visitorcounter,
		}
		tmpl.Execute(w, myvar)
	})

	// Simple health check to integration with Kubernetes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"alive": true}`)
	})

	// log.Fatal(http.ListenAndServe(":8081", nil))
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}

}
