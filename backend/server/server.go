package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultIdleTimeout    = 2 * time.Minute
	defaultReadTimeout    = 10 * time.Second
	defaultWriteTimeout   = 5 * time.Minute
	defaultShutdownPeriod = 30 * time.Second
)

type Application struct {
	subbedFS fs.FS
	Logger   *zap.Logger
}

func NewTelegramApplication() *Application {
	return &Application{}
}

//go:embed static
var staticContent embed.FS

func (app *Application) WebsiteRoutes() *chi.Mux {
	subbedFS, errSubFS := fs.Sub(staticContent, "static")
	if errSubFS != nil {
		panic(fmt.Errorf("failed to create sub fs: %w", errSubFS).Error())
	}
	app.subbedFS = subbedFS

	// http router
	mux := chi.NewRouter()

	mux.NotFound(app.NotFound)
	mux.MethodNotAllowed(app.MethodNotAllowed)

	mux.Use(app.RecoverPanic)

	mux.Use(app.CorsMiddlewareFunc)
	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only consider GET/HEAD requests
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				// Does client accept Brotli?
				if strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
					// Try to open the .br version of the file
					brPath := strings.TrimPrefix(r.URL.Path, "/") + ".br"
					if f, err := app.subbedFS.Open(brPath); err == nil {
						defer f.Close()
						// Set headers
						w.Header().Set("Content-Encoding", "br")
						// Infer MIME type from the original extension
						ext := filepath.Ext(r.URL.Path)
						if typ := mime.TypeByExtension(ext); typ != "" {
							w.Header().Set("Content-Type", typ)
						}
						// Serve the compressed file
						http.ServeContent(w, r, r.URL.Path, time.Now(), f.(io.ReadSeeker))
						return
					}
				}
			}
			// Fallback to normal handling
			next.ServeHTTP(w, r)
		})
	})
	mux.Handle("/*", http.FileServer(http.FS(app.subbedFS)))

	return mux
}

func (app *Application) ServeHTTP(router *chi.Mux, port int) error {

	zapConfig := zap.NewProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	zapConfig.EncoderConfig = encoderConfig

	newLogger, err := zapConfig.Build()
	if err != nil {
		return err
	}
	app.Logger = newLogger

	stdLogger, err := zap.NewStdLogAt(app.Logger, zapcore.WarnLevel)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ErrorLog:     stdLogger,
		IdleTimeout:  defaultIdleTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	shutdownErrorChan := make(chan error)

	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-quitChan

		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownPeriod)
		defer cancel()

		shutdownErrorChan <- srv.Shutdown(ctx)
	}()

	app.Logger.Info("starting server", zap.String("serveraddr", srv.Addr))
	// Glob.Logger.Info("starting server", slog.Group("server", "addr", srv.Addr))

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErrorChan
	if err != nil {
		return err
	}

	app.Logger.Info("starting server", zap.String("serveraddr", srv.Addr))

	return nil
}
