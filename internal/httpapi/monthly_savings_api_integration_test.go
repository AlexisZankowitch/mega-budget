package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

type monthlySavingsResponse struct {
	Year   int32   `json:"year"`
	Months []int32 `json:"months"`
	Values []int64 `json:"values"`
	Total  int64   `json:"total"`
}

func TestMonthlySavingsAnalytics(t *testing.T) {
	categoryA := createTestCategory(t, "Savings-A")
	categoryB := createTestCategory(t, "Savings-B")

	createTestTransaction(t, categoryA.ID, "2031-01-05", -1000, "coffee")
	createTestTransaction(t, categoryA.ID, "2031-01-20", 5000, "salary")
	createTestTransaction(t, categoryB.ID, "2031-02-01", -3000, "books")
	createTestTransaction(t, categoryB.ID, "2031-03-01", 4000, "bonus")

	resp := doRequest(t, http.MethodGet, testServer.URL+"/analytics/monthly-savings?year=2031", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if resp.Header.Get("X-Request-ID") == "" {
		t.Fatalf("missing X-Request-ID header")
	}

	var summary monthlySavingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if summary.Year != 2031 {
		t.Fatalf("year = %d, want 2031", summary.Year)
	}
	if len(summary.Months) != 12 || len(summary.Values) != 12 {
		t.Fatalf("months/values length mismatch: %d/%d", len(summary.Months), len(summary.Values))
	}
	if summary.Values[0] != 4000 {
		t.Fatalf("jan net = %d, want 4000", summary.Values[0])
	}
	if summary.Values[1] != -3000 {
		t.Fatalf("feb net = %d, want -3000", summary.Values[1])
	}
	if summary.Values[2] != 4000 {
		t.Fatalf("mar net = %d, want 4000", summary.Values[2])
	}
	if summary.Total != 5000 {
		t.Fatalf("total = %d, want 5000", summary.Total)
	}
}
