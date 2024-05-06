package http

import (
	"context"
	cp "crypto/rand"
	"github.com/ervitis/poolerchan"
	"io"
	"log"
	"math/big"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/ervitis/codenext/internal/input/core"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type serveAnalyzerPage struct{}

var (
	_, b, _, _ = runtime.Caller(0)
	absPath    = filepath.Join(filepath.Dir(b), "..", "..", "..")
)

func (s serveAnalyzerPage) Handler() http.HandlerFunc {
	type response struct {
		ID string `json:"id"`
	}
	idGenerator := func() string {
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, 16)
		for i := range b {
			n, _ := cp.Int(cp.Reader, big.NewInt(int64(len(charset))))
			b[i] = charset[n.Int64()]
		}
		return string(b)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		idExercise := idGenerator()
		go s.createTask(context.Background(), r.FormValue("exercise"), idExercise)

		http.Redirect(w, r, "/exercise", http.StatusFound)
	}
}

func (s serveAnalyzerPage) createTask(ctx context.Context, code, ID string) {
	type job struct {
		ID   string
		code string
		isOk bool
	}

	queue := poolerchan.NewPoolchan(
		poolerchan.WithNumberOfWorkers(1),
		poolerchan.WithNumberOfJobs(1),
	)

	if err := queue.Queue(func(ctx context.Context) error {
		f, err := os.Create(path.Join(absPath, "exercises", "main.py"))
		if err != nil {
			log.Println(err)
			return err
		}
		_, _ = f.WriteString(code)
		ctnr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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
				Cmd:        []string{"bash", "-c", "/app/executor.sh"},
				WaitingFor: wait.ForLog("done"),
			},
		})
		if err != nil {
			log.Println(err)
			return err
		}
		defer func() {
			if resLog, err := ctnr.Logs(ctx); err != nil {
				log.Println(err)
			} else {
				if logOut, err := os.Create(filepath.Join(absPath, "exercises", "log.out")); err != nil {
					log.Println(err)
				} else {
					_, _ = io.Copy(logOut, resLog)
				}
			}
			_ = ctnr.Terminate(ctx)
		}()
		if err := ctnr.Start(ctx); err != nil {
			log.Println(err)
			return err
		}
		res, err := ctnr.CopyFileFromContainer(ctx, "/app/results/main.out")
		if err != nil {
			log.Println(err)
			return err
		}

		out, err := os.OpenFile(filepath.Join(absPath, "exercises", "main.out"), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			log.Println(err)
			return err
		}
		defer func() {
			_ = out.Close()
		}()
		_, err = io.Copy(out, res)
		if err != nil {
			log.Println(err)
			return err
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go callbackReq(ctx, &wg)
		wg.Wait()
		return nil
	}).Build().Execute(ctx); err != nil {
		log.Println(err)
		return
	}
}

func callbackReq(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8080/"+core.CallbackUrl, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: http.DefaultTransport,
	}

	for i := 0; i < 3; i++ {
		res, err := client.Do(req)
		if err != nil || res.StatusCode > http.StatusInternalServerError {
			time.Sleep(time.Duration(rand.IntN(4)+1) * time.Second)
			continue
		}
		if res.StatusCode > http.StatusOK && res.StatusCode <= http.StatusInternalServerError {
			log.Println(err)
			return
		}
		_ = res.Body.Close()
		break
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
