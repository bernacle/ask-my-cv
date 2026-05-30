package v1

import (
	"github.com/brunosimplicio/ask-my-cv/api/controller/ask"
	"github.com/brunosimplicio/ask-my-cv/api/controller/sources"
	"github.com/go-chi/chi/v5"
)

type Controllers struct {
	Ask     *ask.Controller
	Sources *sources.Controller
}

func Register(r chi.Router, controllers Controllers) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/ask", controllers.Ask.Ask)
		r.Get("/sources", controllers.Sources.List)
	})
}
