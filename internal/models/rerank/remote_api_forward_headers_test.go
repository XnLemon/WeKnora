package rerank

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestOpenAIRerankerForwardsRequestScopedHeaders(t *testing.T) {
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"index":0,"document":{"text":"doc"},"relevance_score":0.9}]}`))
	}))
	defer server.Close()

	reranker, err := NewOpenAIReranker(&RerankerConfig{
		APIKey:    "provider-token",
		BaseURL:   server.URL,
		ModelName: "test-reranker",
		ModelID:   "test-reranker",
	})
	if err != nil {
		t.Fatalf("NewOpenAIReranker: %v", err)
	}
	reranker.SetCustomHeaders(map[string]string{
		"X-Trace-Id": "static-trace",
	})

	ctx := types.WithModelForwardHeaders(context.Background(), map[string]string{
		"X-Trace-Id":    "dynamic-trace",
		"X-Tenant-Id":   "tenant-a",
		"Authorization": "Bearer gateway-token",
	})
	if _, err := reranker.Rerank(ctx, "query", []string{"doc"}); err != nil {
		t.Fatalf("Rerank: %v", err)
	}

	if got := capturedHeaders.Get("X-Trace-Id"); got != "static-trace" {
		t.Fatalf("X-Trace-Id = %q, want static-trace", got)
	}
	if got := capturedHeaders.Get("X-Tenant-Id"); got != "tenant-a" {
		t.Fatalf("X-Tenant-Id = %q, want tenant-a", got)
	}
	if got := capturedHeaders.Get("Authorization"); got != "Bearer gateway-token" {
		t.Fatalf("Authorization = %q, want forwarded gateway token", got)
	}
}
