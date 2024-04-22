package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNewServerExercisePage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "object creation",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServerExercisePage(); got == nil {
				t.Errorf("NewServerExercisePage() = %v, want not nil", got)
			}
		})
	}
}

func Test_serveExercisePage_Handler(t *testing.T) {
	tests := []struct {
		name    string
		checker func(rr *httptest.ResponseRecorder) error
	}{
		{
			name: "handle request",
			checker: func(rr *httptest.ResponseRecorder) error {
				status := rr.Code
				if status != http.StatusOK {
					return errors.New("status is not OK")
				}
				body := rr.Body.String()
				if !(strings.Contains(body, "Codenext submit") ||
					strings.Contains(body, "Write your code test")) {
					return errors.New("response body is not correct")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServerExercisePage()
			rr := httptest.NewRecorder()
			s.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
			if err := tt.checker(rr); err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_serveExercisePage_Method(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "method",
			want: "GET",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServerExercisePage()
			if got := s.Method(); got != tt.want {
				t.Errorf("Method() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serveExercisePage_Middlewares(t *testing.T) {
	tests := []struct {
		name string
		want []func(http.HandlerFunc) http.HandlerFunc
	}{
		{
			name: "middleware",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServerExercisePage()
			if got := s.Middlewares(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Middlewares() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serveExercisePage_Path(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "path",
			want: "/exercise",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServerExercisePage()
			if got := s.Path(); got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}
