package application

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/go-telegram/bot"
)

//go:embed static
var staticContent embed.FS

type Application struct {
	botToken string
	subbedFS fs.FS
	b        *bot.Bot
}

func New(botToken string) (*Application, error) {
	app := &Application{
		botToken: botToken,
	}

	subbedFS, errSubFS := fs.Sub(staticContent, "static")
	if errSubFS != nil {
		return nil, fmt.Errorf("failed to create sub fs: %w", errSubFS)
	}
	app.subbedFS = subbedFS

	b, errBot := bot.New(botToken)
	if errBot != nil {
		return nil, fmt.Errorf("failed to create bot: %w", errBot)
	}

	app.b = b

	return app, nil
}

func (app *Application) Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, ln net.Listener) {
	defer wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/open", app.handlerAPIOpen)
	mux.Handle("/", http.FileServer(http.FS(app.subbedFS)))

	server := http.Server{
		Handler: mux,
	}

	wg.Add(1)
	go func() {
		log.Printf("bot starting")
		app.b.Start(ctx)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Printf("server stopping")
		server.Shutdown(context.Background())
		wg.Done()
	}()

	log.Printf("server started at %s", ln.Addr().String())
	errServe := server.Serve(ln)
	if errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
		log.Printf("error serve: %v", errServe)
	}
	cancel()
}

func (app *Application) handlerAPIOpen(rw http.ResponseWriter, req *http.Request) {
	user, ok := bot.ValidateWebappRequest(req.URL.Query(), app.botToken)
	if !ok {
		http.Error(rw, "unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("%v", user)
}
