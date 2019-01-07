package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/icco/cacophony/models"
)

func main() {
	InitLogging()

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Debugf("Starting up on %s", port)

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatalf("DATABASE_URL is empty!")
	}

	models.InitDB(dbUrl)

	server := http.NewServeMux()
	server.HandleFunc("/", homeHandler)
	server.HandleFunc("/cron", cronHandler)
	server.HandleFunc("/healthz", healthCheckHandler)

	loggedRouter := LoggingMiddleware()(server)

	log.Debugf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, loggedRouter))
}

type healthRespJSON struct {
	Healthy string `json:"healthy"`
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := healthRespJSON{
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
	urls, err := models.SomeSavedUrls(100)
	if err != nil {
		log.WithError(err).Error("Error getting urls")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(urls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func cronHandler(w http.ResponseWriter, r *http.Request) {
	flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
	consumerKey := flags.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flags.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flags.String("access-token", "", "Twitter Access Token")
	accessSecret := flags.String("access-secret", "", "Twitter Access Secret")
	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "TWITTER")

	if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
		log.Fatal("Consumer key/secret and Access token/secret required")
	}

	config := oauth1.NewConfig(*consumerKey, *consumerSecret)
	token := oauth1.NewToken(*accessToken, *accessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	user, resp, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		log.WithError(err).Errorf("Error verifying creds: %+v", resp)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debugf("User: %+v", user.ScreenName)

	// Home Timeline
	homeTimelineParams := &twitter.HomeTimelineParams{
		Count:     200,
		TweetMode: "extended",
	}
	tweets, resp, err := client.Timelines.HomeTimeline(homeTimelineParams)
	if resp.Header.Get("X-Rate-Limit-Remaining") == "0" {
		i, err := strconv.ParseInt(resp.Header.Get("X-Rate-Limit-Reset"), 10, 64)
		if err != nil {
			log.WithError(err).Error("Error converting int")
		}
		tm := time.Unix(i, 0)
		rtlimit := fmt.Errorf("Out of Rate Limit. Returns: %+v", tm)
		http.Error(w, rtlimit.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		log.WithError(err).Error("Error getting tweets: %+v", resp)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, t := range tweets {
		for _, u := range t.Entities.Urls {
			err = models.SaveUrl(u.ExpandedURL, t.IDStr)
			if err != nil {
				log.WithError(err).Error("Error saving url")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	_, err = models.AllSavedUrls()
	if err != nil {
		log.WithError(err).Error("Error getting urls")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(`"ok."`))
}
