package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/gorilla/handlers"
	"golang.org/x/oauth2"
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	server := http.NewServeMux()
	server.HandleFunc("/", homeHandler)
	server.HandleFunc("/cron", cronHandler)
	server.HandleFunc("/_healthcheck.json", healthCheckHandler)

	loggedRouter := handlers.LoggingHandler(os.Stdout, server)

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, loggedRouter))
}

type HealthRespJson struct {
	Healthy string `json:"healthy"`
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := HealthRespJson{
		Healthy: "true",
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(`{"hello": "world"}`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func cronHandler(w http.ResponseWriter, r *http.Request) {
	flags := flag.NewFlagSet("app-auth", flag.ExitOnError)
	accessToken := flags.String("app-access-token", "", "Twitter Application Access Token")
	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "TWITTER")

	if *accessToken == "" {
		log.Fatal("Application Access Token required")
	}

	config := &oauth2.Config{}
	token := &oauth2.Token{AccessToken: *accessToken}
	// OAuth2 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth2.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// user timeline
	userTimelineParams := &twitter.UserTimelineParams{ScreenName: "icco", Count: 2}
	tweets, _, _ = client.Timelines.UserTimeline(userTimelineParams)
	log.Printf("USER TIMELINE:\n%+v\n", tweets)

	w.Write([]byte(`"ok."`))
}
