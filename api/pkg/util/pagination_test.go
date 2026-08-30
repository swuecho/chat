package util

import (
	"net/http/httptest"
	"testing"
)

func TestPaginationParams(t *testing.T) {
	tests := []struct {
		query       string
		wantLimit   int32
		wantOffset  int32
		wantInvalid bool
	}{
		{query: "", wantLimit: 100},
		{query: "?limit=500&offset=20", wantLimit: 500, wantOffset: 20},
		{query: "?limit=0", wantInvalid: true},
		{query: "?limit=501", wantInvalid: true},
		{query: "?limit=nope", wantInvalid: true},
		{query: "?offset=-1", wantInvalid: true},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/"+tt.query, nil)
		limit, offset, err := PaginationParams(req)
		if tt.wantInvalid {
			if err == nil {
				t.Fatalf("PaginationParams(%q) accepted invalid input", tt.query)
			}
			continue
		}
		if err != nil || limit != tt.wantLimit || offset != tt.wantOffset {
			t.Fatalf("PaginationParams(%q) = (%d, %d, %v)", tt.query, limit, offset, err)
		}
	}
}
