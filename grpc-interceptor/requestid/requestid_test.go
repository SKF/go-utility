package requestid

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ExtendNilContext(t *testing.T) {
	newCtx := ExtendContext(nil, "Test_ExtendNilContext")
	req := Extract(newCtx)
	require.Len(t, req.Chain, 1)
	assert.Equal(t, "Test_ExtendNilContext", req.Chain[0])
}

func Test_ExtendBackgroundContext(t *testing.T) {
	newCtx := ExtendContext(context.Background(), "Test_ExtendBackgroundContext")
	req := Extract(newCtx)
	require.Len(t, req.Chain, 1)
	assert.Equal(t, "Test_ExtendBackgroundContext", req.Chain[0])
}

func Test_ExtendExtendedContext(t *testing.T) {
	newCtx1 := ExtendContext(context.Background(), "Test_ExtendExtendedContext_1")
	newCtx2 := outgoingContextWithRequestID(newCtx1, "Test_ExtendExtendedContext_2")
	req := Extract(newCtx2)
	require.Len(t, req.Chain, 1)
	assert.Equal(t, "Test_ExtendExtendedContext_1", req.Chain[0])
}
