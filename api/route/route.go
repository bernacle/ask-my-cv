package route

import (
	"log/slog"
	"net/http"

	"github.com/brunosimplicio/ask-my-cv/api/controller/ask"
	"github.com/brunosimplicio/ask-my-cv/api/controller/health"
	"github.com/brunosimplicio/ask-my-cv/api/controller/sources"
	"github.com/brunosimplicio/ask-my-cv/api/middleware"
	v1 "github.com/brunosimplicio/ask-my-cv/api/route/v1"
	"github.com/brunosimplicio/ask-my-cv/internal/config"
	"github.com/go-chi/chi/v5"
)

type Controllers struct {
	Ask     *ask.Controller
	Health  *health.Controller
	Sources *sources.Controller
}

func New(cfg config.Config, log *slog.Logger, controllers Controllers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer(log))
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	r.Use(middleware.NewRateLimiter(cfg.RateLimitPerMinute, cfg.RateLimitPerDay).Middleware)

	r.Get("/health", controllers.Health.Health)
	v1.Register(r, v1.Controllers{
		Ask:     controllers.Ask,
		Sources: controllers.Sources,
	})

	return r
}
