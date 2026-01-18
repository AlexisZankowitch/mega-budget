package categorieshttp

import (
	"context"
	"database/sql"
	"errors"

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

func (h *Handler) DeleteCategory(ctx context.Context, request api.DeleteCategoryRequestObject) (api.DeleteCategoryResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	err := h.repo.Delete(ctx, request.CategoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.DeleteCategory404JSONResponse{
				Body:    api.Error{Message: "category not found"},
				Headers: api.DeleteCategory404ResponseHeaders{XRequestID: requestID},
			}, nil
		}
		h.logger.Error("delete category: db error", zap.Error(err))
		return nil, err
	}

	h.logger.Info("delete category: deleted", zap.Int64("category_id", request.CategoryId))

	return api.DeleteCategory204Response{
		Headers: api.DeleteCategory204ResponseHeaders{XRequestID: requestID},
	}, nil
}

func (h *Handler) GetCategory(ctx context.Context, request api.GetCategoryRequestObject) (api.GetCategoryResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	cat, err := h.repo.Get(ctx, request.CategoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.GetCategory404JSONResponse{
				Body:    api.Error{Message: "category not found"},
				Headers: api.GetCategory404ResponseHeaders{XRequestID: requestID},
			}, nil
		}
		h.logger.Error("get category: db error", zap.Error(err))
		return nil, err
	}

	return api.GetCategory200JSONResponse{
		Body: api.Category{
			Id:        cat.ID,
			Name:      cat.Name,
			CreatedAt: cat.CreatedAt,
		},
		Headers: api.GetCategory200ResponseHeaders{XRequestID: requestID},
	}, nil
}

func (h *Handler) ListCategories(ctx context.Context, request api.ListCategoriesRequestObject) (api.ListCategoriesResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	cats, err := h.repo.List(ctx)
	if err != nil {
		h.logger.Error("list categories: db error", zap.Error(err))
		return nil, err
	}

	items := make([]api.Category, 0, len(cats))
	for _, c := range cats {
		items = append(items, api.Category{
			Id:        c.ID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt,
		})
	}

	return api.ListCategories200JSONResponse{
		Body:    api.CategoryList{Items: items},
		Headers: api.ListCategories200ResponseHeaders{XRequestID: requestID},
	}, nil
}

func requestIDFromContext(ctx context.Context) string {
	if id, ok := logging.RequestIDFromContext(ctx); ok {
		return id
	}
	return ""
}
