# jsonrpc

[![Travis branch](https://img.shields.io/travis/osamingo/jsonrpc/master.svg)](https://travis-ci.org/osamingo/jsonrpc)
[![codecov](https://codecov.io/gh/osamingo/jsonrpc/branch/master/graph/badge.svg)](https://codecov.io/gh/osamingo/jsonrpc)
[![Test Coverage](https://api.codeclimate.com/v1/badges/e820b394cdbd47103165/test_coverage)](https://codeclimate.com/github/osamingo/jsonrpc/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/osamingo/jsonrpc)](https://goreportcard.com/report/osamingo/jsonrpc)
[![codebeat badge](https://codebeat.co/badges/cbd0290d-200b-4693-80dc-296d9447c35b)](https://codebeat.co/projects/github-com-osamingo-jsonrpc)
[![Maintainability](https://api.codeclimate.com/v1/badges/e820b394cdbd47103165/maintainability)](https://codeclimate.com/github/osamingo/jsonrpc/maintainability)
[![GoDoc](https://godoc.org/github.com/osamingo/jsonrpc?status.svg)](https://godoc.org/github.com/osamingo/jsonrpc)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/osamingo/jsonrpc/master/LICENSE)

## About

- Simple, Poetic, Pithy.
- No `reflect` package.
  - But `reflect` package is used only when invoke the debug handler.
- Support GAE/Go Standard Environment.
- Compliance with [JSON-RPC 2.0](http://www.jsonrpc.org/specification).

Note: If you use Go 1.6, see [v1.0](https://github.com/osamingo/jsonrpc/releases/tag/v1.0).

## Install

```
$ go get -u github.com/osamingo/jsonrpc
```

## Usage

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/intel-go/fastjson"
	"github.com/osamingo/jsonrpc"
)

type (
	EchoHandler struct{}
	EchoParams  struct {
		Name string `json:"name"`
	}
	EchoResult struct {
		Message string `json:"message"`
	}
)

func (h EchoHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var p EchoParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	return EchoResult{
		Message: "Hello, " + p.Name,
	}, nil
}

func main() {

	mr := jsonrpc.NewMethodRepository()

	if err := mr.RegisterMethod("Main.Echo", EchoHandler{}, EchoParams{}, EchoResult{}); err != nil {
		log.Fatalln(err)
	}

	http.Handle("/jrpc", mr)
	http.HandleFunc("/jrpc/debug", mr.ServeDebug)

	if err := http.ListenAndServe(":8080", http.DefaultServeMux); err != nil {
		log.Fatalln(err)
	}
}
```

#### Advanced

```go
package main

import (
	"log"
	"net/http"

	"github.com/osamingo/jsonrpc"
)

type (
	HandleParamsResulter interface {
		jsonrpc.Handler
		Name() string
		Params() interface{}
		Result() interface{}
	}
	Servicer interface {
		MethodName(HandleParamsResulter) string
		Handlers() []HandleParamsResulter
	}
	UserService struct {
		SignUpHandler HandleParamsResulter
		SignInHandler HandleParamsResulter
	}
)

func (us *UserService) MethodName(h HandleParamsResulter) string {
	return "UserService." + h.Name()
}

func (us *UserService) Handlers() []HandleParamsResulter {
	return []HandleParamsResulter{us.SignUpHandler, us.SignInHandler}
}

func NewUserService() *UserService {
	return &UserService{
	// Initialize handlers
	}
}

func main() {

	mr := jsonrpc.NewMethodRepository()

	for _, s := range []Servicer{NewUserService()} {
		for _, h := range s.Handlers() {
			mr.RegisterMethod(s.MethodName(h), h, h.Params(), h.Result())
		}
	}

	http.Handle("/jrpc", mr)
	http.HandleFunc("/jrpc/debug", mr.ServeDebug)

	if err := http.ListenAndServe(":8080", http.DefaultServeMux); err != nil {
		log.Fatalln(err)
	}
}
```

### Result

#### Invoke the Echo method

```
POST /jrpc HTTP/1.1
Accept: application/json, */*
Accept-Encoding: gzip, deflate
Connection: keep-alive
Content-Length: 82
Content-Type: application/json
Host: localhost:8080
User-Agent: HTTPie/0.9.6

{
  "jsonrpc": "2.0",
  "method": "Main.Echo",
  "params": {
    "name": "John Doe"
  },
  "id": "243a718a-2ebb-4e32-8cc8-210c39e8a14b"
}

HTTP/1.1 200 OK
Content-Length: 68
Content-Type: application/json
Date: Mon, 28 Nov 2016 13:48:13 GMT

{
  "jsonrpc": "2.0",
  "result": {
    "message": "Hello, John Doe"
  },
  "id": "243a718a-2ebb-4e32-8cc8-210c39e8a14b"
}
```

#### Access to debug handler

```
GET /jrpc/debug HTTP/1.1
Accept: */*
Accept-Encoding: gzip, deflate
Connection: keep-alive
Host: localhost:8080
User-Agent: HTTPie/0.9.6



HTTP/1.1 200 OK
Content-Length: 408
Content-Type: application/json
Date: Mon, 28 Nov 2016 13:56:24 GMT

[
  {
    "handler": "EchoHandler",
    "name": "Main.Echo",
    "params": {
      "$ref": "#/definitions/EchoParams",
      "$schema": "http://json-schema.org/draft-04/schema#",
      "definitions": {
        "EchoParams": {
          "additionalProperties": false,
          "properties": {
            "name": {
              "type": "string"
            }
          },
          "required": [
            "name"
          ],
          "type": "object"
        }
      }
    },
    "result": {
      "$ref": "#/definitions/EchoResult",
      "$schema": "http://json-schema.org/draft-04/schema#",
      "definitions": {
        "EchoResult": {
          "additionalProperties": false,
          "properties": {
            "message": {
              "type": "string"
            }
          },
          "required": [
            "message"
          ],
          "type": "object"
        }
      }
    }
  }
]
```

## License

Released under the [MIT License](https://github.com/osamingo/jsonrpc/blob/master/LICENSE).
