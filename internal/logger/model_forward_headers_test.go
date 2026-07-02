package logger

import (
	"context"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestCloneContextPreservesModelForwardHeaders(t *testing.T) {
	ctx := types.WithModelForwardHeaders(context.Background(), map[string]string{
		"X-Trace-Id": "trace-1",
	})

	cloned := CloneContext(ctx)

	assert.Equal(t, map[string]string{"X-Trace-Id": "trace-1"}, types.ModelForwardHeadersFromContext(cloned))
}
