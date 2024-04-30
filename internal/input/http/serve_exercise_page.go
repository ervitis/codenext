package http

import (
	"html/template"
	"net/http"

	"github.com/ervitis/codenext/internal/input/core"
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
	unresolvedExerciseCode := `def main():
	# Write your code here
	print("hello")

if __name__ == "__main__":
	main()`

	html := `
<html>
    <head>
    <title>Codenext submit</title>
    </head>
    <body>
        <h2 style="text-align: center;">Codenext</h2>
        <form action="/resolve" method="POST">
            <textarea cols="100" rows="50" name="exercise">
{{.ExerciseCode}}
</textarea>
            <button type="submit">Solve exercise</button>
        </form>
    </body>
</html>
`
	type data struct {
		ExerciseCode string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("exercise").Parse(html)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = tmpl.Execute(w, data{ExerciseCode: unresolvedExerciseCode})
	}
}

func NewServerExercisePage() core.Endpoint {
	return &serveExercisePage{}
}
