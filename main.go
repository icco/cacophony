package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/coreos/pkg/flagutil"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icco/cacophony/models"
	"github.com/icco/cacophony/workers"
	"github.com/icco/gutil/logging"
	"github.com/icco/gutil/otel"
	"go.uber.org/zap"
)

var (
	service = "cacophony"
	project = "icco-cloud"
	log     = logging.Must(logging.NewLogger(service))

	flgs                  = flag.NewFlagSet("default", flag.ExitOnError)
	twitterConsumerKey    = flgs.String("twitter-consumer-key", "", "Twitter Consumer Key")
	twitterConsumerSecret = flgs.String("twitter-consumer-secret", "", "Twitter Consumer Secret")
	twitterAccessToken    = flgs.String("twitter-access-token", "", "Twitter Access Token")
	twitterAccessSecret   = flgs.String("twitter-access-secret", "", "Twitter Access Secret")
	mastoServer           = flgs.String("mastodon-server", "https://merveilles.town", "Mastodon server")
	mastoClientID         = flgs.String("mastodon-client-id", "", "Mastodon Client ID")
	mastoClientSecret     = flgs.String("mastodon-client-secret", "", "Mastodon Client Secret")
	mastoAccessToken      = flgs.String("mastodon-access-token", "", "Mastodon Access Token")
	port                  = flgs.Int("port", 8080, "Server local port")
	dbURL                 = flgs.String("database-url", "", "Postgres database url")
)

func init() {
	flgs.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flgs, "")
}

func main() {
	log.Infow("Starting up", "host", fmt.Sprintf("http://localhost:%d", *port))

	if *dbURL == "" {
		log.Fatal("DATABASE_URL is empty!")
	}

	ctx := context.Background()
	if err := otel.Init(ctx, log, project, service); err != nil {
		log.Errorw("could not init opentelemetry", zap.Error(err))
	}

	models.InitDB(*dbURL)

	r := chi.NewRouter()
	r.Use(otel.Middleware)
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

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
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
}

func cronHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := workers.Twitter(ctx, *twitterConsumerKey, *twitterConsumerSecret, *twitterAccessToken, *twitterAccessSecret); err != nil {
		log.Errorw("Error getting tweets", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := workers.Mastodon(ctx, *mastoServer, *mastoClientID, *mastoClientSecret, *mastoAccessToken); err != nil {
		log.Errorw("Error getting toots", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(`"ok."`))
}
