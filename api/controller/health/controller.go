package health

import (
	"net/http"

	"github.com/brunosimplicio/ask-my-cv/api/response"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
