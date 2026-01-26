package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

type summaryRow struct {
	CategoryID int64   `json:"category_id"`
	Values     []int64 `json:"values"`
	Total      int64   `json:"total"`
	Average    int64   `json:"average"`
}

type summarySection struct {
	Rows         []summaryRow `json:"rows"`
	ColumnTotals []int64      `json:"column_totals"`
	Total        int64        `json:"total"`
}

type summaryResponse struct {
	Year     int32          `json:"year"`
	Months   []int32        `json:"months"`
	Spending summarySection `json:"spending"`
	Income   summarySection `json:"income"`
}

func TestTransactionsSummaryAnalytics(t *testing.T) {
	categoryA := createTestCategory(t, "Analytics-A")
	categoryB := createTestCategory(t, "Analytics-B")

	createTestTransaction(t, categoryA.ID, "2030-01-05", -1000, "coffee")
	createTestTransaction(t, categoryA.ID, "2030-01-20", 2000, "refund")
	createTestTransaction(t, categoryB.ID, "2030-02-01", -3000, "books")
	createTestTransaction(t, categoryB.ID, "2030-03-01", 4000, "salary")

	resp := doRequest(t, http.MethodGet, testServer.URL+"/analytics/transactions-summary?year=2030", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if resp.Header.Get("X-Request-ID") == "" {
		t.Fatalf("missing X-Request-ID header")
	}

	var summary summaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if summary.Year != 2030 {
		t.Fatalf("year = %d, want 2030", summary.Year)
	}
	if len(summary.Months) != 12 {
		t.Fatalf("months length = %d, want 12", len(summary.Months))
	}
	if summary.Months[0] != 1 || summary.Months[11] != 12 {
		t.Fatalf("months should be 1..12, got %v", summary.Months)
	}

	spendingA, ok := findSummaryRow(summary.Spending.Rows, categoryA.ID)
	if !ok || len(spendingA.Values) != 12 || spendingA.Values[0] != 1000 {
		t.Fatalf("spending for categoryA: %+v", spendingA)
	}
	if spendingA.Average != 83 {
		t.Fatalf("spending avg for categoryA = %d, want 83", spendingA.Average)
	}
	spendingB, ok := findSummaryRow(summary.Spending.Rows, categoryB.ID)
	if !ok || len(spendingB.Values) != 12 || spendingB.Values[1] != 3000 {
		t.Fatalf("spending for categoryB: %+v", spendingB)
	}
	if spendingB.Average != 250 {
		t.Fatalf("spending avg for categoryB = %d, want 250", spendingB.Average)
	}
	if summary.Spending.Total != 4000 {
		t.Fatalf("spending total = %d, want 4000", summary.Spending.Total)
	}
	if len(summary.Spending.ColumnTotals) != 12 || summary.Spending.ColumnTotals[0] != 1000 || summary.Spending.ColumnTotals[1] != 3000 {
		t.Fatalf("spending column totals = %v", summary.Spending.ColumnTotals)
	}

	incomeA, ok := findSummaryRow(summary.Income.Rows, categoryA.ID)
	if !ok || len(incomeA.Values) != 12 || incomeA.Values[0] != 2000 {
		t.Fatalf("income for categoryA: %+v", incomeA)
	}
	if incomeA.Average != 166 {
		t.Fatalf("income avg for categoryA = %d, want 166", incomeA.Average)
	}
	incomeB, ok := findSummaryRow(summary.Income.Rows, categoryB.ID)
	if !ok || len(incomeB.Values) != 12 || incomeB.Values[2] != 4000 {
		t.Fatalf("income for categoryB: %+v", incomeB)
	}
	if incomeB.Average != 333 {
		t.Fatalf("income avg for categoryB = %d, want 333", incomeB.Average)
	}
	if summary.Income.Total != 6000 {
		t.Fatalf("income total = %d, want 6000", summary.Income.Total)
	}
	if len(summary.Income.ColumnTotals) != 12 || summary.Income.ColumnTotals[0] != 2000 || summary.Income.ColumnTotals[2] != 4000 {
		t.Fatalf("income column totals = %v", summary.Income.ColumnTotals)
	}
}

func createTestCategory(t *testing.T, name string) categoryResponse {
	t.Helper()

	body := []byte(`{"name":"` + name + `"}`)
	resp := doRequest(t, http.MethodPost, testServer.URL+"/categories", body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var created categoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode category: %v", err)
	}
	return created
}

func createTestTransaction(t *testing.T, categoryID int64, date string, amountCents int64, description string) {
	t.Helper()

	body := []byte(`{"transaction_date":"` + date + `","amount_cents":` + itoa(amountCents) + `,"category_id":` + itoa(categoryID) + `,"description":"` + description + `"}`)
	resp := doRequest(t, http.MethodPost, testServer.URL+"/transactions", body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func findSummaryRow(rows []summaryRow, categoryID int64) (summaryRow, bool) {
	for _, row := range rows {
		if row.CategoryID == categoryID {
			return row, true
		}
	}
	return summaryRow{}, false
}
