package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

type categoryResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type categoryListResponse struct {
	Items []categoryResponse `json:"items"`
}

func TestCategoriesHTTPCRUD(t *testing.T) {
	// Subtests share state and must not be run in isolation or parallel.
	var created categoryResponse

	t.Run("create category", func(t *testing.T) {
		name := "Category-" + time.Now().UTC().Format("150405.000000000")
		body := []byte(`{"name":"` + name + `"}`)
		resp := doRequest(t, http.MethodPost, testServer.URL+"/categories", body)
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

	t.Run("get category", func(t *testing.T) {
		resp := doRequest(t, http.MethodGet, testServer.URL+"/categories/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("list categories", func(t *testing.T) {
		resp := doRequest(t, http.MethodGet, testServer.URL+"/categories", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var list categoryListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(list.Items) == 0 {
			t.Fatalf("expected at least one category")
		}
	})

	t.Run("update category", func(t *testing.T) {
		body := []byte(`{"name":"Updated Category"}`)
		resp := doRequest(t, http.MethodPut, testServer.URL+"/categories/"+itoa(created.ID), body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("delete category", func(t *testing.T) {
		resp := doRequest(t, http.MethodDelete, testServer.URL+"/categories/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", resp.StatusCode)
		}
	})

	t.Run("get after delete returns not found", func(t *testing.T) {
		resp := doRequest(t, http.MethodGet, testServer.URL+"/categories/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", resp.StatusCode)
		}
	})
}
