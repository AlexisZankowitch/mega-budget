package httpserver_test

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
