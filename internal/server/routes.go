package server

import (
	"database/sql"
	"net/http"

	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Routes(app *config.AppConfig, db *sql.DB) http.Handler {
	metricHandler := handler.NewMetricHandler(app, db)
	mux := chi.NewRouter()
	mux.Use(GzipHandle)
	mux.Mount("/debug", middleware.Profiler())

	mux.Post("/update/{type}/{name}/{value}", metricHandler.Update)

	if app.Parameters.CryptoKey != "" {
		decoder := NewDecoderMiddleware(app.Parameters.CryptoKey)
		mux.Group(func(r chi.Router) {
			r.Use(decoder.Decode)
			r.Post("/update/", metricHandler.UpdateJSON)
			r.Post("/update", metricHandler.UpdateJSON)
			r.Post("/updates/", metricHandler.UpdateBatch)
		})
	} else {
		mux.Post("/update/", metricHandler.UpdateJSON)
		mux.Post("/update", metricHandler.UpdateJSON)
		mux.Post("/updates/", metricHandler.UpdateBatch)
	}

	mux.Post("/value/", metricHandler.GetJSONValue)
	mux.Post("/value", metricHandler.GetJSONValue)

	mux.Get("/value/{type}/{name}", metricHandler.GetValue)
	mux.Get("/", metricHandler.Get)
	mux.Get("/ping", metricHandler.Ping)

	return mux
}
