package jsonrpc

import (
	"context"
	"testing"

	"github.com/intel-go/fastjson"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {

	c := context.Background()
	id := fastjson.RawMessage("1")
	c = WithRequestID(c, &id)
	var pick *fastjson.RawMessage
	require.NotPanics(t, func() {
		pick = RequestID(c)
	})
	require.Equal(t, &id, pick)
}
