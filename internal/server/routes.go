package server

import (
	"database/sql"
	"net/http"

	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/handler"
	"github.com/go-chi/chi"
)

func Routes(app *config.AppConfig, db *sql.DB) http.Handler {
	metricHandler := handler.NewMetricHandler(app, db)
	mux := chi.NewRouter()
	mux.Use(GzipHandle)
	mux.Post("/update/{type}/{name}/{value}", metricHandler.Update)
	mux.Post("/update/", metricHandler.UpdateJSON)
	mux.Post("/update", metricHandler.UpdateJSON)
	mux.Post("/value/", metricHandler.GetJSONValue)
	mux.Post("/value", metricHandler.GetJSONValue)
	mux.Get("/value/{type}/{name}", metricHandler.GetValue)
	mux.Get("/", metricHandler.Get)
	mux.Get("/ping", metricHandler.Ping)
	return mux
}
