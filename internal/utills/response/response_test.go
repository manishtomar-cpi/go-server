package response_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/manishtomar-cpi/go-server/internal/utills/response"
)

func TestWriteJson(t *testing.T) {
	t.Parallel() // allow this test to run in parallel with others

	type testCase struct {
		name            string
		status          int
		data            any
		wantErr         bool
		wantContentType string
		assertBody      func(t *testing.T, body string)
	}

	tests := []testCase{
		{
			name:            "encodes_struct_and_sets_headers",
			status:          200,
			data:            response.Response{Status: response.StatusOk, Error: ""},
			wantErr:         false,
			wantContentType: "application/json",
			assertBody: func(t *testing.T, body string) {
				// decode and assert deterministically
				var got map[string]any
				if err := json.Unmarshal([]byte(body), &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["Status"] != response.StatusOk {
					t.Fatalf("want Status=%q, got=%v", response.StatusOk, got["Status"])
				}
			},
		},
		{
			name:            "encodes_map_payload",
			status:          201,
			data:            map[string]any{"id": 123},
			wantErr:         false,
			wantContentType: "application/json",
			assertBody: func(t *testing.T, body string) {
				var got map[string]any
				if err := json.Unmarshal([]byte(body), &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["id"] != float64(123) { // json numbers decode as float64
					t.Fatalf("want id=123, got=%v", got["id"])
				}
			},
		},
		{
			name:            "returns_error_when_json_encoding_fails",
			status:          500,
			data:            make(chan int), // json cannot encode channels
			wantErr:         true,
			wantContentType: "application/json",
			assertBody: func(t *testing.T, body string) {
				// When encoding fails, encoder returns error and should not write a body
				if strings.TrimSpace(body) != "" {
					t.Fatalf("expected empty body on encode error, got: %q", body)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()

			err := response.WriteJson(rr, tc.status, tc.data)

			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// status code
			if rr.Code != tc.status {
				t.Fatalf("status mismatch: want %d, got %d", tc.status, rr.Code)
			}

			// content type
			if ct := rr.Header().Get("Content-Type"); ct != tc.wantContentType {
				t.Fatalf("content-type mismatch: want %q, got %q", tc.wantContentType, ct)
			}

			// body assertions
			tc.assertBody(t, rr.Body.String())
		})
	}
}
