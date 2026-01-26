package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

type transactionResponse struct {
	ID              int64   `json:"id"`
	TransactionDate string  `json:"transaction_date"`
	CategoryID      *int64  `json:"category_id"`
	AmountCents     int64   `json:"amount_cents"`
	Description     *string `json:"description"`
	CreatedAt       string  `json:"created_at"`
}

type transactionListResponse struct {
	Items []transactionResponse `json:"items"`
}

func TestTransactionsHTTPCRUD(t *testing.T) {
	// Subtests share state and must not be run in isolation or parallel.
	var created transactionResponse
	var second transactionResponse

	t.Run("create transaction", func(t *testing.T) {
		body := []byte(`{"transaction_date":"2026-02-01","amount_cents":-1257,"description":"lunch"}`)
		resp := doRequest(t, http.MethodPost, testServer.URL+"/transactions", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}
		if resp.Header.Get("X-Request-ID") == "" {
			t.Fatalf("missing X-Request-ID header")
		}

		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if created.ID == 0 {
			t.Fatalf("expected id to be set")
		}
	})

	t.Run("create second transaction", func(t *testing.T) {
		body := []byte(`{"transaction_date":"2026-01-31","amount_cents":-500,"description":"snack"}`)
		resp := doRequest(t, http.MethodPost, testServer.URL+"/transactions", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&second); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if second.ID == 0 {
			t.Fatalf("expected id to be set")
		}
	})

	t.Run("get transaction", func(t *testing.T) {
		resp := doRequest(t, http.MethodGet, testServer.URL+"/transactions/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("list transactions", func(t *testing.T) {
		resp := doRequest(t, http.MethodGet, testServer.URL+"/transactions", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var list transactionListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(list.Items) == 0 {
			t.Fatalf("expected at least one transaction")
		}
	})

	t.Run("list transactions with cursor", func(t *testing.T) {
		resp := doRequest(t, http.MethodGet, testServer.URL+"/transactions?limit=1", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var first transactionListResponse
		if err := json.NewDecoder(resp.Body).Decode(&first); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(first.Items) != 1 {
			t.Fatalf("expected 1 transaction, got %d", len(first.Items))
		}

		cursorDate := first.Items[0].TransactionDate
		cursorID := first.Items[0].ID
		url := testServer.URL + "/transactions?after_date=" + cursorDate + "&after_id=" + itoa(cursorID)
		resp = doRequest(t, http.MethodGet, url, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var next transactionListResponse
		if err := json.NewDecoder(resp.Body).Decode(&next); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(next.Items) < 1 {
			t.Fatalf("expected at least 1 transaction after cursor, got %d", len(next.Items))
		}
		if next.Items[0].ID == cursorID {
			t.Fatalf("expected next item to differ from cursor")
		}
	})

	t.Run("update transaction", func(t *testing.T) {
		body := []byte(`{"transaction_date":"2026-02-02","amount_cents":-2000,"description":"updated"}`)
		resp := doRequest(t, http.MethodPut, testServer.URL+"/transactions/"+itoa(created.ID), body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("delete transaction", func(t *testing.T) {
		resp := doRequest(t, http.MethodDelete, testServer.URL+"/transactions/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", resp.StatusCode)
		}
	})

	t.Run("get after delete returns not found", func(t *testing.T) {
		resp := doRequest(t, http.MethodGet, testServer.URL+"/transactions/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", resp.StatusCode)
		}
	})
}

func TestTransactionsListFilters(t *testing.T) {
	category := createTestCategory(t, "Filter-Category")

	createTestTransaction(t, category.ID, "2032-05-10", -1200, "may-spending")
	createTestTransaction(t, category.ID, "2032-05-12", 5000, "may-income")
	createTestTransaction(t, category.ID, "2032-04-20", -700, "april-spending")

	t.Run("filters by month/category/spending", func(t *testing.T) {
		url := testServer.URL + "/transactions?from_date=2032-05-01&to_date=2032-05-31&category_id=" + itoa(category.ID) + "&type=spending&limit=50"
		resp := doRequest(t, http.MethodGet, url, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var list transactionListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(list.Items) != 1 {
			t.Fatalf("expected 1 transaction, got %d", len(list.Items))
		}
		if list.Items[0].Description == nil || *list.Items[0].Description != "may-spending" {
			t.Fatalf("unexpected transaction: %+v", list.Items[0])
		}
	})

	t.Run("filters by income type", func(t *testing.T) {
		url := testServer.URL + "/transactions?from_date=2032-05-01&to_date=2032-05-31&category_id=" + itoa(category.ID) + "&type=income&limit=50"
		resp := doRequest(t, http.MethodGet, url, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var list transactionListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(list.Items) != 1 {
			t.Fatalf("expected 1 transaction, got %d", len(list.Items))
		}
		if list.Items[0].Description == nil || *list.Items[0].Description != "may-income" {
			t.Fatalf("unexpected transaction: %+v", list.Items[0])
		}
	})

	t.Run("start_date aliases to to_date", func(t *testing.T) {
		url := testServer.URL + "/transactions?start_date=2032-05-11&category_id=" + itoa(category.ID) + "&limit=50"
		resp := doRequest(t, http.MethodGet, url, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var list transactionListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(list.Items) == 0 {
			t.Fatalf("expected at least 1 transaction")
		}
		for _, item := range list.Items {
			if item.TransactionDate > "2032-05-11" {
				t.Fatalf("unexpected transaction after start_date: %+v", item)
			}
		}
	})
}
