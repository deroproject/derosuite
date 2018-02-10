package jsonrpc

import "github.com/intel-go/fastjson"

// Unmarshal decodes JSON-RPC params.
func Unmarshal(params *fastjson.RawMessage, dst interface{}) *Error {
	if params == nil {
		return ErrInvalidParams()
	}
	if err := fastjson.Unmarshal(*params, dst); err != nil {
		return ErrInvalidParams()
	}
	return nil
}
