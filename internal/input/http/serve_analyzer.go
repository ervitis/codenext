package http

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/api/types/container"
	"github.com/ervitis/codenext/internal/input/core"
	"github.com/testcontainers/testcontainers-go"
)

type serveAnalyzerPage struct{}

var (
	_, b, _, _ = runtime.Caller(0)
	absPath    = filepath.Join(filepath.Dir(b), "..", "..", "..")
)

func (s serveAnalyzerPage) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rawCode := r.FormValue("exercise")
		f, err := os.Create(path.Join(absPath, "exercises", "main.py"))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = f.WriteString(rawCode)
		ctnr, err := testcontainers.GenericContainer(r.Context(), testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:      "docker.io/library/python:3.12",
				WorkingDir: "/app",
				Files: []testcontainers.ContainerFile{
					{
						HostFilePath:      f.Name(),
						ContainerFilePath: "/app/main.py",
						FileMode:          0o400,
					},
					{
						HostFilePath:      filepath.Join(absPath, "exercises", "executor.sh"),
						FileMode:          0o700,
						ContainerFilePath: "/app/executor.sh",
					},
				},
				HostConfigModifier: func(cfg *container.HostConfig) {
					cfg.Binds = append(cfg.Binds, filepath.Join(absPath, "exercises")+":/app/results")
				},
				Cmd: []string{"bash", "-c", "/app/executor.sh"},
			},
		})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if resLog, err := ctnr.Logs(r.Context()); err != nil {
				log.Println(err)
			} else {
				if logOut, err := os.Create(filepath.Join(absPath, "exercises", "log.out")); err != nil {
					log.Println(err)
				} else {
					_, _ = io.Copy(logOut, resLog)
				}
			}
			_ = ctnr.Terminate(r.Context())
		}()
		if err := ctnr.Start(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		res, err := ctnr.CopyFileFromContainer(r.Context(), "/app/results/main.out")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}

		out, err := os.OpenFile(filepath.Join(absPath, "exercises", "main.out"), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		defer func() {
			_ = out.Close()
		}()
		_, err = io.Copy(out, res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		r.PostForm.Set("exercise", rawCode)

		http.Redirect(w, r, "/exercise", http.StatusFound)
	}
}

func (s serveAnalyzerPage) Middlewares() []func(http.HandlerFunc) http.HandlerFunc {
	return nil
}

func (s serveAnalyzerPage) Path() string {
	return "/resolve"
}

func (s serveAnalyzerPage) Method() string {
	return http.MethodPost
}

func NewServerAnalyzerPage() core.Endpoint {
	return &serveAnalyzerPage{}
}
