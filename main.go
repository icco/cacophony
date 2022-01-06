package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icco/cacophony/models"
	"github.com/icco/cron/shared"
	"github.com/icco/cron/tweets"
	"github.com/icco/gutil/logging"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

var (
	service = "cacophony"
	project = "icco-cloud"
	log     = logging.Must(logging.NewLogger(service))
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Infow("Starting up", "host", fmt.Sprintf("http://localhost:%s", port))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is empty!")
	}

	if os.Getenv("ENABLE_STACKDRIVER") != "" {
		labels := &stackdriver.Labels{}
		labels.Set("app", service, "The name of the current app.")
		sd, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID:               project,
			MonitoredResource:       monitoredresource.Autodetect(),
			DefaultMonitoringLabels: labels,
			DefaultTraceAttributes:  map[string]interface{}{"app": service},
		})

		if err != nil {
			log.Fatalw("failed to create the stackdriver exporter", zap.Error(err))
		}
		defer sd.Flush()

		view.RegisterExporter(sd)
		trace.RegisterExporter(sd)
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		})
	}

	models.InitDB(dbURL)

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(logging.Middleware(log.Desugar(), project))

	r.Get("/", homeHandler)
	r.Get("/cron", cronHandler)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok."))
		if err != nil {
			log.Errorw("could not write response", zap.Error(err))
		}
	})
	h := &ochttp.Handler{
		Handler: r,
	}
	if err := view.Register([]*view.View{
		ochttp.ServerRequestCountView,
		ochttp.ServerResponseCountByStatusCode,
	}...); err != nil {
		log.Fatalw("Failed to register ochttp views", zap.Error(err))
	}

	log.Fatal(http.ListenAndServe(":"+port, h))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	cntStr := r.URL.Query().Get("count")
	cnt := 100
	if cntStr != "" {
		i, err := strconv.Atoi(cntStr)
		if err != nil {
			log.Errorw("Error parsing count", zap.Error(err))
		} else {
			cnt = i
		}
	}

	urls, err := models.SomeSavedURLs(r.Context(), cnt)
	if err != nil {
		log.Errorw("Error getting urls", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		log.Errorw("Error encoding json", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return
}

func cronHandler(w http.ResponseWriter, r *http.Request) {
	flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
	consumerKey := flags.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flags.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flags.String("access-token", "", "Twitter Access Token")
	accessSecret := flags.String("access-secret", "", "Twitter Access Secret")
	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "TWITTER")

	ctx := r.Context()

	if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
		log.Fatal("Consumer key/secret and Access token/secret required")
	}

	config := oauth1.NewConfig(*consumerKey, *consumerSecret)
	token := oauth1.NewToken(*accessToken, *accessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(ctx, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	user, resp, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		log.Errorw("Error verifying creds", "response", resp, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debugf("User: %+v", user.ScreenName)

	// Home Timeline
	homeTimelineParams := &twitter.HomeTimelineParams{
		Count:     200,
		TweetMode: "extended",
	}
	homeTweets, resp, err := client.Timelines.HomeTimeline(homeTimelineParams)
	if resp.Header.Get("X-Rate-Limit-Remaining") == "0" {
		i, err := strconv.ParseInt(resp.Header.Get("X-Rate-Limit-Reset"), 10, 64)
		if err != nil {
			log.Errorw("Error converting int", zap.Error(err))
		}
		tm := time.Unix(i, 0)
		rtlimit := fmt.Errorf("out of Rate Limit. Returns: %+v", tm)
		http.Error(w, rtlimit.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		log.Errorw("Error getting tweets", "response", resp, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := tweets.Twitter{
		Config: shared.Config{
			Log: log,
		},
		GraphQLToken: os.Getenv("GQL_TOKEN"),
	}

	for _, t := range homeTweets {
		// Save tweet to DB via graphql
		err := c.UploadTweet(ctx, t)
		if err != nil {
			log.Errorw("problem uploading tweet", zap.Error(err))
		}

		for _, u := range t.Entities.Urls {
			err = models.SaveURL(ctx, u.ExpandedURL, t.IDStr)
			if err != nil {
				log.Errorw("Error saving url", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(`"ok."`))
}
