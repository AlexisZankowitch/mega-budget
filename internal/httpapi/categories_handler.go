package httpapi

import (
	"context"
	"database/sql"
	"errors"

	"go.uber.org/zap"

	"zankowitch.com/go-db-app/internal/api"
	"zankowitch.com/go-db-app/internal/categories"
)

type CategoriesHandler struct {
	repo   *categories.Repository
	logger *zap.Logger
}

func NewCategoriesHandler(repo *categories.Repository, logger *zap.Logger) *CategoriesHandler {
	return &CategoriesHandler{repo: repo, logger: logger}
}

func (h *CategoriesHandler) CreateCategory(ctx context.Context, request api.CreateCategoryRequestObject) (api.CreateCategoryResponseObject, error) {
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

func (h *CategoriesHandler) DeleteCategory(ctx context.Context, request api.DeleteCategoryRequestObject) (api.DeleteCategoryResponseObject, error) {
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

func (h *CategoriesHandler) GetCategory(ctx context.Context, request api.GetCategoryRequestObject) (api.GetCategoryResponseObject, error) {
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

func (h *CategoriesHandler) ListCategories(ctx context.Context, request api.ListCategoriesRequestObject) (api.ListCategoriesResponseObject, error) {
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

func (h *CategoriesHandler) UpdateCategory(ctx context.Context, request api.UpdateCategoryRequestObject) (api.UpdateCategoryResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	logger := h.logger.With(zap.String("request_id", requestID))
	if request.Body == nil {
		logger.Warn("update category: missing request body")
		return api.UpdateCategory400JSONResponse{
			Body:    api.Error{Message: "missing request body"},
			Headers: api.UpdateCategory400ResponseHeaders{XRequestID: requestID},
		}, nil
	}

	updated, err := h.repo.Update(ctx, request.CategoryId, categories.UpdateInput{Name: request.Body.Name})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.UpdateCategory404JSONResponse{
				Body:    api.Error{Message: "category not found"},
				Headers: api.UpdateCategory404ResponseHeaders{XRequestID: requestID},
			}, nil
		}
		logger.Error("update category: db error", zap.Error(err))
		return nil, err
	}

	return api.UpdateCategory200JSONResponse{
		Body: api.Category{
			Id:        updated.ID,
			Name:      updated.Name,
			CreatedAt: updated.CreatedAt,
		},
		Headers: api.UpdateCategory200ResponseHeaders{XRequestID: requestID},
	}, nil
}
