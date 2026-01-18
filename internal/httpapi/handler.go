package httpapi

import (
	"context"

	"zankowitch.com/go-db-app/internal/api"
	categorieshttp "zankowitch.com/go-db-app/internal/categories/http"
	transactionshttp "zankowitch.com/go-db-app/internal/transactions/http"
)

type Handler struct {
	transactions *transactionshttp.Handler
	categories   *categorieshttp.Handler
}

func NewHandler(transactions *transactionshttp.Handler, categories *categorieshttp.Handler) *Handler {
	return &Handler{transactions: transactions, categories: categories}
}

func (h *Handler) CreateTransaction(ctx context.Context, request api.CreateTransactionRequestObject) (api.CreateTransactionResponseObject, error) {
	return h.transactions.CreateTransaction(ctx, request)
}

func (h *Handler) DeleteTransaction(ctx context.Context, request api.DeleteTransactionRequestObject) (api.DeleteTransactionResponseObject, error) {
	return h.transactions.DeleteTransaction(ctx, request)
}

func (h *Handler) GetTransaction(ctx context.Context, request api.GetTransactionRequestObject) (api.GetTransactionResponseObject, error) {
	return h.transactions.GetTransaction(ctx, request)
}

func (h *Handler) UpdateTransaction(ctx context.Context, request api.UpdateTransactionRequestObject) (api.UpdateTransactionResponseObject, error) {
	return h.transactions.UpdateTransaction(ctx, request)
}

func (h *Handler) ListTransactions(ctx context.Context, request api.ListTransactionsRequestObject) (api.ListTransactionsResponseObject, error) {
	return h.transactions.ListTransactions(ctx, request)
}

func (h *Handler) CreateCategory(ctx context.Context, request api.CreateCategoryRequestObject) (api.CreateCategoryResponseObject, error) {
	return h.categories.CreateCategory(ctx, request)
}

var _ api.StrictServerInterface = (*Handler)(nil)
