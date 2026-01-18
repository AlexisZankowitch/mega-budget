package transactionshttp

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"

	"zankowitch.com/go-db-app/internal/api"
	"zankowitch.com/go-db-app/internal/logging"
	"zankowitch.com/go-db-app/internal/transactions"
)

type Handler struct {
	repo   *transactions.Repository
	logger *zap.Logger
}

func NewHandler(repo *transactions.Repository, logger *zap.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) CreateTransaction(ctx context.Context, request api.CreateTransactionRequestObject) (api.CreateTransactionResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	logger := h.logger.With(zap.String("request_id", requestID))
	if request.Body == nil {
		logger.Warn("create transaction: missing request body")
		return api.CreateTransaction400JSONResponse{
			Body:    api.Error{Message: "missing request body"},
			Headers: api.CreateTransaction400ResponseHeaders{XRequestID: requestID},
		}, nil
	}

	logger.Debug(
		"create transaction: request",
		zap.String("transaction_date", request.Body.TransactionDate.String()),
		zap.Any("category_id", request.Body.CategoryId),
		zap.Int64("amount_cents", request.Body.AmountCents),
		zap.String("description", stringPtrValue(request.Body.Description)),
	)

	created, err := h.repo.Create(ctx, transactions.CreateInput{
		TransactionDate: request.Body.TransactionDate.Time,
		CategoryID:      request.Body.CategoryId,
		AmountCents:     request.Body.AmountCents,
		Description:     request.Body.Description,
	})
	if err != nil {
		logger.Error("create transaction: db error", zap.Error(err))
		return nil, err
	}

	logger.Info("create transaction: created", zap.Int64("transaction_id", created.ID))

	response := api.Transaction{
		Id:              created.ID,
		TransactionDate: types.Date{Time: created.TransactionDate},
		CategoryId:      created.CategoryID,
		AmountCents:     created.AmountCents,
		Description:     created.Description,
		CreatedAt:       created.CreatedAt,
	}

	return api.CreateTransaction201JSONResponse{
		Body:    response,
		Headers: api.CreateTransaction201ResponseHeaders{XRequestID: requestID},
	}, nil
}

func (h *Handler) DeleteTransaction(ctx context.Context, request api.DeleteTransactionRequestObject) (api.DeleteTransactionResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	err := h.repo.Delete(ctx, request.TransactionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.DeleteTransaction404JSONResponse{
				Body:    api.Error{Message: "transaction not found"},
				Headers: api.DeleteTransaction404ResponseHeaders{XRequestID: requestID},
			}, nil
		}
		h.logger.Error("delete transaction: db error", zap.Error(err))
		return nil, err
	}

	h.logger.Info("delete transaction: deleted", zap.Int64("transaction_id", request.TransactionId))

	return api.DeleteTransaction204Response{
		Headers: api.DeleteTransaction204ResponseHeaders{XRequestID: requestID},
	}, nil
}

func (h *Handler) ListTransactions(ctx context.Context, request api.ListTransactionsRequestObject) (api.ListTransactionsResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	logger := h.logger.With(zap.String("request_id", requestID))

	limit := int32(50)
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	var startDate *time.Time
	if request.Params.StartDate != nil {
		start := request.Params.StartDate.Time
		startDate = &start
	}

	hasAfterDate := request.Params.AfterDate != nil
	hasAfterID := request.Params.AfterId != nil
	hasCursor := hasAfterDate && hasAfterID
	if hasAfterDate != hasAfterID {
		return api.ListTransactions400JSONResponse{
			Body:    api.Error{Message: "after_date and after_id must be provided together"},
			Headers: api.ListTransactions400ResponseHeaders{XRequestID: requestID},
		}, nil
	}

	var afterDate *time.Time
	var afterID *int64
	if hasCursor {
		after := request.Params.AfterDate.Time
		afterDate = &after
		afterID = request.Params.AfterId
	}

	rows, err := h.repo.ListAfter(ctx, int(limit), startDate, afterDate, afterID)
	if err != nil {
		logger.Error("list transactions: db error", zap.Error(err))
		return nil, err
	}

	items := make([]api.Transaction, 0, len(rows))
	for _, row := range rows {
		items = append(items, api.Transaction{
			Id:              row.ID,
			TransactionDate: types.Date{Time: row.TransactionDate},
			CategoryId:      row.CategoryID,
			AmountCents:     row.AmountCents,
			Description:     row.Description,
			CreatedAt:       row.CreatedAt,
		})
	}

	return api.ListTransactions200JSONResponse{
		Body:    api.TransactionList{Items: items},
		Headers: api.ListTransactions200ResponseHeaders{XRequestID: requestID},
	}, nil
}

func (h *Handler) GetTransaction(ctx context.Context, request api.GetTransactionRequestObject) (api.GetTransactionResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	row, err := h.repo.Get(ctx, request.TransactionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.GetTransaction404JSONResponse{
				Body:    api.Error{Message: "transaction not found"},
				Headers: api.GetTransaction404ResponseHeaders{XRequestID: requestID},
			}, nil
		}
		h.logger.Error("get transaction: db error", zap.Error(err))
		return nil, err
	}

	response := api.Transaction{
		Id:              row.ID,
		TransactionDate: types.Date{Time: row.TransactionDate},
		CategoryId:      row.CategoryID,
		AmountCents:     row.AmountCents,
		Description:     row.Description,
		CreatedAt:       row.CreatedAt,
	}

	return api.GetTransaction200JSONResponse{
		Body:    response,
		Headers: api.GetTransaction200ResponseHeaders{XRequestID: requestID},
	}, nil
}

func (h *Handler) UpdateTransaction(ctx context.Context, request api.UpdateTransactionRequestObject) (api.UpdateTransactionResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	logger := h.logger.With(zap.String("request_id", requestID))
	if request.Body == nil {
		logger.Warn("update transaction: missing request body")
		return api.UpdateTransaction400JSONResponse{
			Body:    api.Error{Message: "missing request body"},
			Headers: api.UpdateTransaction400ResponseHeaders{XRequestID: requestID},
		}, nil
	}

	updated, err := h.repo.Update(ctx, request.TransactionId, transactions.UpdateInput{
		TransactionDate: request.Body.TransactionDate.Time,
		CategoryID:      request.Body.CategoryId,
		AmountCents:     request.Body.AmountCents,
		Description:     request.Body.Description,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.UpdateTransaction404JSONResponse{
				Body:    api.Error{Message: "transaction not found"},
				Headers: api.UpdateTransaction404ResponseHeaders{XRequestID: requestID},
			}, nil
		}
		logger.Error("update transaction: db error", zap.Error(err))
		return nil, err
	}

	response := api.Transaction{
		Id:              updated.ID,
		TransactionDate: types.Date{Time: updated.TransactionDate},
		CategoryId:      updated.CategoryID,
		AmountCents:     updated.AmountCents,
		Description:     updated.Description,
		CreatedAt:       updated.CreatedAt,
	}

	return api.UpdateTransaction200JSONResponse{
		Body:    response,
		Headers: api.UpdateTransaction200ResponseHeaders{XRequestID: requestID},
	}, nil
}

var _ api.StrictServerInterface = (*Handler)(nil)

func stringPtrValue(v *string) string {
	if v == nil {
		return "<nil>"
	}
	return *v
}

func requestIDFromContext(ctx context.Context) string {
	if id, ok := logging.RequestIDFromContext(ctx); ok {
		return id
	}
	return ""
}
