package core

import "net/http"

type Endpoint interface {
	Handler() http.HandlerFunc
	Middlewares() []func(http.HandlerFunc) http.HandlerFunc
	Path() string
	Method() string
}
