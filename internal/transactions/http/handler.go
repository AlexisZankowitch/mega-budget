package transactionshttp

import (
	"context"

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
