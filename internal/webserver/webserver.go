package webserver

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"go.uber.org/zap"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/webserver/handler"
	mw "genshin-quiz/internal/webserver/middleware"
)

type Server struct {
	router     *chi.Mux
	serverAddr string
}

func NewServer(app *config.App) *Server {
	// Setup router
	r := chi.NewRouter()

	// Sentry 中间件 - 用于自动错误跟踪
	if app.Config.SentryDSN != "" {
		sentryHandler := sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		})
		r.Use(sentryHandler.Handle)
	}

	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.Logger(app.Logger))

	// 使用自定义的错误处理中间件，替代 chi 的 Recoverer
	r.Use(mw.Handler(app.Logger))

	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint - 必须在 OpenAPI 路由之前定义，避免被覆盖
	r.Route("/health", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(
				[]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`),
			)
			if err != nil {
				app.Logger.Error("Failed to write health check response", zap.Error(err))
			}
		})
		r.Head("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})
	})

	// Setup API routes without authentication (temporarily for testing)
	// TODO: Add JWT authentication back for protected routes
	// r.Group(func(r chi.Router) {
	// 	r.Use(jwtauth.Verifier(app.JWTAuth))
	// 	r.Use(mw.Authenticator)
	baseURL := ""
	serverOptions := oapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  mw.HandleBadRequestError(app),
		ResponseErrorHandlerFunc: mw.HandleResponseErrorWithLog(app),
	}
	strictHandler := oapi.NewStrictHandlerWithOptions(
		handler.NewHandler(app),
		[]oapi.StrictMiddlewareFunc{},
		serverOptions,
	)
	oapi.HandlerFromMuxWithBaseURL(strictHandler, r, baseURL)

	defer sentry.Flush(2 * time.Second)

	return &Server{
		router:     r,
		serverAddr: fmt.Sprintf("%s:%s", app.Server.Host, app.Server.Port),
	}
}

func (s *Server) Start() {
	log.Print("Starting server", zap.String("addr", s.serverAddr))

	const maxHeaderBytes = 1 << 20
	const readTimeout = 10 * time.Second
	const writeTimeout = 30 * time.Second
	const idleTimeout = 10 * time.Second

	// Create HTTP server
	srv := &http.Server{
		Addr:           s.serverAddr,
		Handler:        s.router,
		ReadTimeout:    readTimeout,
		MaxHeaderBytes: maxHeaderBytes,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}

func (s *Server) Router() chi.Router {
	return s.router
}
