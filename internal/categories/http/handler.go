package categorieshttp

import (
	"context"

	"go.uber.org/zap"

	"zankowitch.com/go-db-app/internal/api"
	"zankowitch.com/go-db-app/internal/categories"
	"zankowitch.com/go-db-app/internal/logging"
)

type Handler struct {
	repo   *categories.Repository
	logger *zap.Logger
}

func NewHandler(repo *categories.Repository, logger *zap.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) CreateCategory(ctx context.Context, request api.CreateCategoryRequestObject) (api.CreateCategoryResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	logger := h.logger.With(zap.String("request_id", requestID))
	if request.Body == nil {
		logger.Warn("create category: missing request body")
		return api.CreateCategory400JSONResponse{
			Body:    api.Error{Message: "missing request body"},
			Headers: api.CreateCategory400ResponseHeaders{XRequestID: requestID},
		}, nil
	}

	created, err := h.repo.Create(ctx, categories.CreateInput{Name: request.Body.Name})
	if err != nil {
		logger.Error("create category: db error", zap.Error(err))
		return nil, err
	}

	logger.Info("create category: created", zap.Int64("category_id", created.ID))

	return api.CreateCategory201JSONResponse{
		Body: api.Category{
			Id:        created.ID,
			Name:      created.Name,
			CreatedAt: created.CreatedAt,
		},
		Headers: api.CreateCategory201ResponseHeaders{XRequestID: requestID},
	}, nil
}

func requestIDFromContext(ctx context.Context) string {
	if id, ok := logging.RequestIDFromContext(ctx); ok {
		return id
	}
	return ""
}
