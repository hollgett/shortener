package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hollgett/shortener.git/internal/config"
	"github.com/hollgett/shortener.git/internal/handlers"
	"github.com/hollgett/shortener.git/internal/logger"
	"github.com/hollgett/shortener.git/internal/service"
	"github.com/hollgett/shortener.git/internal/store"
	"github.com/hollgett/shortener.git/internal/worker"
	"go.uber.org/zap"
)

// timeout for graceful shutdown
const (
	shutDownPeriod     = 15 * time.Second
	shutDownHardPeriod = 5 * time.Second
)

type App struct {
	logger       *logger.Logger
	cfg          *config.ShortenerConfig
	store        store.Store
	service      *service.Service
	workerDelete *worker.DeleteWorker
	handlers     *handlers.Handlers
	middleware   *handlers.Middleware
}

func NewApp() *App {
	return &App{}
}

// start application with graceful shutdown, build router, logger
func (a *App) Run() {
	rootCtx, rootStop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer rootStop()

	// get logger
	a.setLogger()
	defer a.logger.Close()

	//get config args
	a.cfg = config.NewConfig()

	//get layers
	a.setLayers()

	// get server with base context and routers
	ongoingCtx, stopOngoingCtx := context.WithCancel(context.Background())
	server := a.buildServer(ongoingCtx)

	go func() {
		a.logger.Info("Starting server", zap.String("address", a.cfg.Addr), zap.String("static address", a.cfg.BaseURL))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-rootCtx.Done()
	a.logger.Info("shutdown app...")
	rootStop()

	shutDownCtx, cancel := context.WithTimeout(context.Background(), shutDownPeriod)
	defer cancel()

	err := server.Shutdown(shutDownCtx)
	stopOngoingCtx()
	if err != nil {
		time.Sleep(shutDownHardPeriod)
	}

	a.workerDelete.ShutDown()

	err = a.store.Close()
	if err != nil {
		a.logger.Info("store close", zap.Error(err))
	}

	a.logger.Info("app is shutdown")
}

// set logger to app
func (a *App) setLogger() {
	logger, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	a.logger = logger
}

// set layers store, service, handler
func (a *App) setLayers() {
	//get service
	var err error
	if a.store, err = store.NewStore(a.logger, a.cfg.FilePath, a.cfg.DatabaseDSN); err != nil {
		panic(err)
	}

	//get worker
	a.workerDelete = worker.NewDeleteWorker(a.logger, a.store)
	go a.workerDelete.Run()

	//get service
	a.service = service.NewService(a.logger, a.store, a.workerDelete.DeleteCh)

	//get handlers
	a.handlers = handlers.NewHandlers(a.logger, a.service, a.cfg.BaseURL)

	//get middleware
	a.middleware = handlers.NewMiddleware(a.logger, a.cfg.SecretKey)
}

// creates server with config and routes
func (a *App) buildServer(ongoingCtx context.Context) *http.Server {
	return &http.Server{
		Addr:    a.cfg.Addr,
		Handler: a.setRouter(),
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}
}

// creates paths for handlers
func (a *App) setRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", a.handlers.CreateOrRedirectText)
	mux.HandleFunc("/ping", a.handlers.PingDatabase)
	mux.HandleFunc("/api/shorten", a.handlers.CreateAPIShortURL)
	mux.HandleFunc("/api/shorten/batch", a.handlers.CreateAPIShortURLs)
	mux.HandleFunc("/api/user/urls", a.handlers.ControllerUserURLs)
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(handlers.UserKeyCtx)
		val, ok := userID.(string)
		if !ok {
			http.Error(w, "failed get user id", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(val))
	})

	return handlers.ConveyorMiddleware(mux,
		a.middleware.AuthMiddleware,
		a.middleware.RequestLogged,
		a.middleware.UnCompress,
		a.middleware.Compress,
		a.middleware.ResponseLogged,
	)
}
