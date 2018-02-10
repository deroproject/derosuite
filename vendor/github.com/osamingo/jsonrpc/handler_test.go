package jsonrpc

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/intel-go/fastjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type handler struct {
	F func(c context.Context, params *fastjson.RawMessage) (interface{}, *Error)
}

func (h *handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *Error) {
	return h.F(c, params)
}

func TestHandler(t *testing.T) {

	mr := NewMethodRepository()

	rec := httptest.NewRecorder()
	r, err := http.NewRequest("", "", nil)
	require.NoError(t, err)

	mr.ServeHTTP(rec, r)

	res := Response{}
	err = fastjson.NewDecoder(rec.Body).Decode(&res)
	require.NoError(t, err)
	assert.NotNil(t, res.Error)

	rec = httptest.NewRecorder()
	r, err = http.NewRequest("", "", bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":"test","method":"hello","params":{}}`)))
	require.NoError(t, err)
	r.Header.Set("Content-Type", "application/json")

	mr.ServeHTTP(rec, r)
	res = Response{}
	err = fastjson.NewDecoder(rec.Body).Decode(&res)
	require.NoError(t, err)
	assert.NotNil(t, res.Error)

	h1 := &handler{}
	h1.F = func(c context.Context, params *fastjson.RawMessage) (interface{}, *Error) {
		return "hello", nil
	}
	require.NoError(t, mr.RegisterMethod("hello", h1, nil, nil))
	h2 := &handler{}
	h2.F = func(c context.Context, params *fastjson.RawMessage) (interface{}, *Error) {
		return nil, ErrInternal()
	}
	require.NoError(t, mr.RegisterMethod("bye", h2, nil, nil))

	rec = httptest.NewRecorder()
	r, err = http.NewRequest("", "", bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":"test","method":"hello","params":{}}`)))
	require.NoError(t, err)
	r.Header.Set("Content-Type", "application/json")

	mr.ServeHTTP(rec, r)
	res = Response{}
	err = fastjson.NewDecoder(rec.Body).Decode(&res)
	require.NoError(t, err)
	assert.Nil(t, res.Error)
	assert.Equal(t, "hello", res.Result)

	rec = httptest.NewRecorder()
	r, err = http.NewRequest("", "", bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":"test","method":"bye","params":{}}`)))
	require.NoError(t, err)
	r.Header.Set("Content-Type", "application/json")

	mr.ServeHTTP(rec, r)
	res = Response{}
	err = fastjson.NewDecoder(rec.Body).Decode(&res)
	require.NoError(t, err)
	assert.NotNil(t, res.Error)
}
