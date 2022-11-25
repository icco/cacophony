package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icco/cacophony/models"
	"github.com/icco/gutil/logging"
	"github.com/icco/gutil/otel"
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

	ctx := context.Background()
	if err := otel.Init(ctx, log, project, service); err != nil {
		log.Errorw("could not init opentelemetry", zap.Error(err))
	}

	models.InitDB(dbURL)

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

	log.Fatal(http.ListenAndServe(":"+port, r))
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

	if err := twitterCronWorker(ctx); err != nil {
		log.Errorw("Error getting tweets", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(`"ok."`))
}
