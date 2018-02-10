package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	var err interface{} = &Error{}
	_, ok := err.(error)
	require.True(t, ok)
}

func TestError_Error(t *testing.T) {

	err := &Error{
		Code:    ErrorCode(100),
		Message: "test",
		Data: map[string]string{
			"test": "test",
		},
	}

	assert.Equal(t, "jsonrpc: code: 100, message: test, data: map[test:test]", err.Error())
}

func TestErrParse(t *testing.T) {
	err := ErrParse()
	require.Equal(t, ErrorCodeParse, err.Code)
}

func TestErrInvalidRequest(t *testing.T) {
	err := ErrInvalidRequest()
	require.Equal(t, ErrorCodeInvalidRequest, err.Code)
}

func TestErrMethodNotFound(t *testing.T) {
	err := ErrMethodNotFound()
	require.Equal(t, ErrorCodeMethodNotFound, err.Code)
}

func TestErrInvalidParams(t *testing.T) {
	err := ErrInvalidParams()
	require.Equal(t, ErrorCodeInvalidParams, err.Code)
}

func TestErrInternal(t *testing.T) {
	err := ErrInternal()
	require.Equal(t, ErrorCodeInternal, err.Code)
}
