package api

import (
	"fmt"
	"net/http"

	"insider-messaging/docs"
	"insider-messaging/internal/application"
	"insider-messaging/internal/config"
	"insider-messaging/internal/domain/repository"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter HTTP router'ı oluşturur ve tüm endpoint'leri tanımlar
func NewRouter(sched application.SchedulerController, repo repository.MessageRepository, cfg *config.Config) http.Handler {
	h := NewHandler(sched, repo, cfg)
	r := mux.NewRouter()

	apiKeyMiddleware := APIKeyMiddleware(cfg)

	api := r.PathPrefix("/api").Subrouter()
	api.Use(apiKeyMiddleware)
	api.HandleFunc("/auto", h.StartStop).Methods("POST", "GET")
	api.HandleFunc("/sent", h.ListSent).Methods("GET")
	api.HandleFunc("/messages", h.CreateMessage).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", cfg.Port)
	docs.SwaggerInfo.BasePath = "/api"
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%s/swagger/doc.json", cfg.Port)),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	return r
}

// NewServer HTTP server'ı oluşturur
func NewServer(cfg *config.Config, handler http.Handler) *http.Server {
	addr := fmt.Sprintf(":%s", cfg.Port)
	return &http.Server{Addr: addr, Handler: handler}
}
