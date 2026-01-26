package httpapi

import (
	"context"

	"go.uber.org/zap"

	"zankowitch.com/go-db-app/internal/api"
	"zankowitch.com/go-db-app/internal/categories"
	"zankowitch.com/go-db-app/internal/transactions"
)

type AnalyticsHandler struct {
	txRepo  *transactions.Repository
	catRepo *categories.Repository
	logger  *zap.Logger
}

func NewAnalyticsHandler(txRepo *transactions.Repository, catRepo *categories.Repository, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{txRepo: txRepo, catRepo: catRepo, logger: logger}
}

func (h *AnalyticsHandler) GetTransactionsSummary(ctx context.Context, request api.GetTransactionsSummaryRequestObject) (api.GetTransactionsSummaryResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	year := int(request.Params.Year)
	if year <= 0 {
		return api.GetTransactionsSummary400JSONResponse{
			Body:    api.Error{Message: "year must be a positive integer"},
			Headers: api.GetTransactionsSummary400ResponseHeaders{XRequestID: requestID},
		}, nil
	}

	categoriesList, err := h.catRepo.List(ctx)
	if err != nil {
		h.logger.Error("transactions summary: list categories failed", zap.Error(err))
		return nil, err
	}

	spendingRows, err := h.txRepo.ListMonthlySpendingByCategory(ctx, year)
	if err != nil {
		h.logger.Error("transactions summary: spending query failed", zap.Error(err))
		return nil, err
	}

	incomeRows, err := h.txRepo.ListMonthlyIncomeByCategory(ctx, year)
	if err != nil {
		h.logger.Error("transactions summary: income query failed", zap.Error(err))
		return nil, err
	}

	months := make([]int32, 12)
	for i := range months {
		months[i] = int32(i + 1)
	}

	spending := buildSummarySection(categoriesList, spendingRows)
	income := buildSummarySection(categoriesList, incomeRows)

	return api.GetTransactionsSummary200JSONResponse{
		Body: api.TransactionsSummary{
			Year:     request.Params.Year,
			Months:   months,
			Spending: spending,
			Income:   income,
		},
		Headers: api.GetTransactionsSummary200ResponseHeaders{XRequestID: requestID},
	}, nil
}

func (h *AnalyticsHandler) GetMonthlySavings(ctx context.Context, request api.GetMonthlySavingsRequestObject) (api.GetMonthlySavingsResponseObject, error) {
	requestID := requestIDFromContext(ctx)
	year := int(request.Params.Year)
	if year <= 0 {
		return api.GetMonthlySavings400JSONResponse{
			Body:    api.Error{Message: "year must be a positive integer"},
			Headers: api.GetMonthlySavings400ResponseHeaders{XRequestID: requestID},
		}, nil
	}

	netRows, err := h.txRepo.ListMonthlyNetTotals(ctx, year)
	if err != nil {
		h.logger.Error("monthly savings: query failed", zap.Error(err))
		return nil, err
	}

	months := make([]int32, 12)
	values := make([]int64, 12)
	for i := range months {
		months[i] = int32(i + 1)
	}

	for _, row := range netRows {
		if row.Month < 1 || row.Month > 12 {
			continue
		}
		values[row.Month-1] = row.AmountCents
	}

	var total int64
	for _, v := range values {
		total += v
	}

	return api.GetMonthlySavings200JSONResponse{
		Body: api.MonthlySavings{
			Year:   request.Params.Year,
			Months: months,
			Values: values,
			Total:  total,
		},
		Headers: api.GetMonthlySavings200ResponseHeaders{XRequestID: requestID},
	}, nil
}

func buildSummarySection(categoriesList []categories.Category, rows []transactions.MonthlyCategoryTotal) api.TransactionsSummarySection {
	valuesByCategory := make(map[int64][]int64, len(categoriesList))
	for _, c := range categoriesList {
		valuesByCategory[c.ID] = make([]int64, 12)
	}

	for _, row := range rows {
		values, ok := valuesByCategory[row.CategoryID]
		if !ok {
			continue
		}
		if row.Month < 1 || row.Month > 12 {
			continue
		}
		values[row.Month-1] = row.AmountCents
	}

	rowsOut := make([]api.TransactionsSummaryRow, 0, len(categoriesList))
	columnTotals := make([]int64, 12)
	var grandTotal int64

	for _, c := range categoriesList {
		values := valuesByCategory[c.ID]
		var total int64
		for i, v := range values {
			total += v
			columnTotals[i] += v
		}
		rowsOut = append(rowsOut, api.TransactionsSummaryRow{
			CategoryId: c.ID,
			Values:     values,
			Total:      total,
		})
		grandTotal += total
	}

	return api.TransactionsSummarySection{
		Rows:         rowsOut,
		ColumnTotals: columnTotals,
		Total:        grandTotal,
	}
}
