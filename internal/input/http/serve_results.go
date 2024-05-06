package http

import (
	"github.com/ervitis/codenext/internal/input/core"
	"net/http"
)

type serverResultsPage struct{}

func (s serverResultsPage) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/exercise", http.StatusFound)
	}
}

func (s serverResultsPage) Middlewares() []func(http.HandlerFunc) http.HandlerFunc {
	return nil
}

func (s serverResultsPage) Path() string {
	return core.CallbackUrl
}

func (s serverResultsPage) Method() string {
	return http.MethodPost
}

func NewServerResultsPage() core.Endpoint {
	return &serverResultsPage{}
}
