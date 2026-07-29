package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBatchRequestAcceptsCompleteDocument(t *testing.T) {
	var req batchRequest
	data := []byte(`{"ops":[{"op":"add","url":"https://example.com","title":"Example","label":"reference","metadata":{"source":"export"}}]}`)
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatal(err)
	}
	if len(req.Ops) != 1 {
		t.Fatalf("operation count = %d, want 1", len(req.Ops))
	}
	op := req.Ops[0]
	if op.Op != batchOpAdd || op.URL != "https://example.com" || op.Title != "Example" {
		t.Fatalf("unexpected operation: %#v", op)
	}
	if op.Label != "reference" {
		t.Fatalf("label = %q, want reference", op.Label)
	}
	if op.Metadata["source"] != "export" {
		t.Fatalf("metadata source = %v, want export", op.Metadata["source"])
	}
}

func TestServeBatchReportsOversizedRequest(t *testing.T) {
	body := `{"ops":[{"op":"add","url":"https://example.com","text":"` +
		strings.Repeat("x", maxBatchBodySize) + `"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/batch", strings.NewReader(body))
	rec := httptest.NewRecorder()

	serveBatch(&webContext{Request: req, Response: rec})

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
	var response batchResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if want := "request body exceeds the 5 MiB limit"; response.Error != want {
		t.Fatalf("error = %q, want %q", response.Error, want)
	}
}

func TestServeBatchReportsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/batch", strings.NewReader(`{"ops":`))
	rec := httptest.NewRecorder()

	serveBatch(&webContext{Request: req, Response: rec})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var response batchResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error != "invalid JSON" {
		t.Fatalf("error = %q, want %q", response.Error, "invalid JSON")
	}
}
