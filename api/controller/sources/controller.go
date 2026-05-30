package sources

import (
	"context"
	"net/http"

	"github.com/brunosimplicio/ask-my-cv/api/response"
	domaindocument "github.com/brunosimplicio/ask-my-cv/internal/domain/document"
	sourcedto "github.com/brunosimplicio/ask-my-cv/internal/dto/source"
)

type Repository interface {
	ListSources(ctx context.Context) ([]domaindocument.Source, error)
}

type Controller struct {
	repository Repository
}

func New(repository Repository) *Controller {
	return &Controller{repository: repository}
}

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	sources, err := c.repository.ListSources(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	payload := sourcedto.ListResponse{Sources: make([]sourcedto.Source, 0, len(sources))}
	for _, item := range sources {
		payload.Sources = append(payload.Sources, sourcedto.Source{
			ID:             item.ID.String(),
			SourcePath:     item.SourcePath,
			Title:          item.Title,
			EmbeddingModel: item.EmbeddingModel,
			ChunkCount:     item.ChunkCount,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	response.JSON(w, http.StatusOK, payload)
}
