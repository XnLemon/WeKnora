package types

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelForwardHeadersContextRoundTripCopiesMap(t *testing.T) {
	ctx := WithModelForwardHeaders(context.Background(), map[string]string{
		"X-Trace-Id": "trace-1",
		"X-User-Id":  "user-1",
	})

	got := ModelForwardHeadersFromContext(ctx)
	assert.Equal(t, map[string]string{
		"X-Trace-Id": "trace-1",
		"X-User-Id":  "user-1",
	}, got)

	got["X-Trace-Id"] = "mutated"
	assert.Equal(t, "trace-1", ModelForwardHeadersFromContext(ctx)["X-Trace-Id"])
}

func TestApplyModelForwardHeadersToRequestPreservesExistingNonReservedHeaders(t *testing.T) {
	ctx := WithModelForwardHeaders(context.Background(), map[string]string{
		"X-Trace-Id":  "trace-dynamic",
		"X-Tenant-Id": "tenant-a",
	})
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("X-Trace-Id", "trace-static")

	ApplyModelForwardHeadersToRequest(ctx, req)

	assert.Equal(t, "trace-static", req.Header.Get("X-Trace-Id"))
	assert.Equal(t, "tenant-a", req.Header.Get("X-Tenant-Id"))
}

func TestApplyModelForwardHeadersToRequestOverridesReservedHeaders(t *testing.T) {
	ctx := WithModelForwardHeaders(context.Background(), map[string]string{
		"Authorization": "Bearer gateway-token",
	})
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer provider-token")

	ApplyModelForwardHeadersToRequest(ctx, req)

	assert.Equal(t, "Bearer gateway-token", req.Header.Get("Authorization"))
}

func TestWrapHTTPClientWithModelForwardHeadersInjectsFromRequestContext(t *testing.T) {
	var gotTrace string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTrace = r.Header.Get("X-Trace-Id")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := WrapHTTPClientWithModelForwardHeaders(nil)
	ctx := WithModelForwardHeaders(context.Background(), map[string]string{"X-Trace-Id": "trace-rt"})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "trace-rt", gotTrace)
}
