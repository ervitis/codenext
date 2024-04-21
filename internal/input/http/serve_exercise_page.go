package http

import (
	"github.com/ervitis/codenext/internal/input/core"
	"html/template"
	"net/http"
)

type serveExercisePage struct{}

func (s serveExercisePage) Middlewares() []func(http.HandlerFunc) http.HandlerFunc {
	return nil
}

func (s serveExercisePage) Path() string {
	return "/exercise"
}

func (s serveExercisePage) Method() string {
	return http.MethodGet
}

func (s serveExercisePage) Handler() http.HandlerFunc {
	html := `
<html>
    <head>
    <title></title>
    </head>
    <body>
        <h2 style="text-align: center;">Codenext</h2>
        <form action="/resolve" method="POST">
            <textarea cols="100" rows="50" name="exercise">Write your code test</textarea>
            <button type="submit">Solve exercise</button>
        </form>
    </body>
</html>
`
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("exercise").Parse(html)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = tmpl.Execute(w, nil)
	}
}

func NewServerExercisePage() core.Endpoint {
	return &serveExercisePage{}
}
