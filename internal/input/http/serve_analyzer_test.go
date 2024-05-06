package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewServerAnalyzerPage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "page instance",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServerAnalyzerPage(); got == nil {
				t.Errorf("NewServerAnalyzerPage() = %v is nil", got)
			}
		})
	}
}

func Test_serveAnalyzerPage_Handler(t *testing.T) {
	tests := []struct {
		name    string
		checker func(rr *httptest.ResponseRecorder) error
	}{
		{
			name: "page handler",
			checker: func(rr *httptest.ResponseRecorder) error {
				if rr.Code != http.StatusAccepted {
					return fmt.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusAccepted)
				}
				return nil
			},
		},
		{
			name: "create container and file to handle result",
			checker: func(rr *httptest.ResponseRecorder) error {
				if rr.Code != http.StatusAccepted {
					return fmt.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusAccepted)
				}
				// check file with retrying
				var allErrs error
				for i := 0; i < 3; i++ {
					info, terr := os.Stat(filepath.Join(absPath, "exercises", "main.py"))
					if terr != nil && os.IsNotExist(terr) {
						time.Sleep(1 * time.Second)
					}
					if terr != nil {
						time.Sleep(1 * time.Second)
					}
					f, _ := os.Open(filepath.Join(absPath, "exercises", info.Name()))
					defer f.Close()
					b, _ := io.ReadAll(f)
					if !strings.Contains(string(b), "hello test") {
						return fmt.Errorf("file should contains hello test")
					}
					if terr != nil {
						allErrs = errors.Join(terr)
					} else {
						allErrs = nil
						break
					}
				}

				return allErrs
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServerAnalyzerPage()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/resolve", nil)
			pythonCode := `def main():
	# Write your code here
	print("hello test")

if __name__ == "__main__":
	main()`
			req.Form = url.Values{
				"exercise": []string{pythonCode},
			}
			s.Handler().ServeHTTP(rr, req)
			if err := tt.checker(rr); err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_serveAnalyzerPage_Method(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "method",
			want: http.MethodPost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServerAnalyzerPage()
			if got := s.Method(); got != tt.want {
				t.Errorf("Method() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serveAnalyzerPage_Middlewares(t *testing.T) {
	tests := []struct {
		name string
		want []func(http.HandlerFunc) http.HandlerFunc
	}{
		{
			name: "middlewares",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServerAnalyzerPage()
			if got := s.Middlewares(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Middlewares() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serveAnalyzerPage_Path(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "get page path",
			want: "/resolve",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServerAnalyzerPage()
			if got := s.Path(); got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}
