package webserver

import (
	"fmt"
	"genshin-quiz/internal/authkeys"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

// Mode controls which routes this process serves.
// ModeAll keeps the original single-process behaviour; ModeAuth and ModeQuiz
// let the same codebase run as two independent services behind the gateway.
type Mode string

const (
	ModeAll  Mode = "all"
	ModeAuth Mode = "auth"
	ModeQuiz Mode = "quiz"
)

func NewServer(app *config.App) *Server {
	return NewServerWithMode(app, ModeAll)
}

// initAuthKeys 准备 token 签名 / 校验密钥：auth-service 持有私钥，其余服务只拉取公钥。
func initAuthKeys(app *config.App, mode Mode) {
	if mode == ModeQuiz {
		jwksURL := os.Getenv("AUTH_JWKS_URL")
		if jwksURL == "" {
			app.Logger.Fatal("AUTH_JWKS_URL is required when SERVICE_MODE=quiz")
		}
		authkeys.InitRemoteVerifier(jwksURL)
		return
	}
	keyFile := os.Getenv("JWT_KEY_FILE")
	if keyFile == "" {
		keyFile = "./data/jwt-es256.pem"
	}
	created, err := authkeys.InitSigner(keyFile)
	if err != nil {
		app.Logger.Fatal("Failed to initialize JWT signing key", zap.Error(err))
	}
	if created {
		app.Logger.Warn("Generated a new JWT signing key", zap.String("file", keyFile))
	}
}

func NewServerWithMode(app *config.App, mode Mode) *Server {
	initAuthKeys(app, mode)

	// Setup router
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(mw.SecureRealIP)
	r.Use(mw.Logger(app.Logger))
	// 使用自定义的错误处理中间件，替代 chi 的 Recoverer
	r.Use(mw.Handler(app))
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS：身份服务（/auth/*）依赖会话 Cookie，只允许受信任站点并启用 credentials；
	// 其余业务接口仍使用 Bearer token，保持对任意来源开放。
	openCORS := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	authCORS := cors.Handler(cors.Options{
		AllowOriginFunc:  func(_ *http.Request, origin string) bool { return handler.IsTrustedOrigin(origin) },
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	// 站外应用（已登记的 SSO 客户端）只允许跨域读取授权码兑换接口，且不带 Cookie
	ssoTokenCORS := cors.Handler(cors.Options{
		AllowOriginFunc: func(_ *http.Request, origin string) bool {
			return handler.IsTrustedOrigin(origin) || handler.IsSSOClientOrigin(origin)
		},
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
		MaxAge:         300,
	})
	r.Use(func(next http.Handler) http.Handler {
		open, strict, token := openCORS(next), authCORS(next), ssoTokenCORS(next)
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch {
			case req.URL.Path == "/auth/sso/token":
				token.ServeHTTP(w, req)
			case req.URL.Path == "/auth" || strings.HasPrefix(req.URL.Path, "/auth/"):
				strict.ServeHTTP(w, req)
			default:
				open.ServeHTTP(w, req)
			}
		})
	})
	// service 层通过 context 取得响应写入器，用于下发 / 清除会话 Cookie
	r.Use(mw.WithResponseWriter)

	// Route isolation: when running as auth-service or quiz-service, requests
	// outside the process's responsibility fail fast instead of being handled
	// by the other domain's handlers.
	if mode != ModeAll {
		r.Use(routeModeFilter(mode))
	}

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

	h := handler.NewHandler(app)

	r.Get("/auth/check", h.EmailCheck)
	r.Get("/auth/.well-known/jwks.json", h.JWKS)
	r.Post("/auth/session/refresh", h.SessionRefresh)
	r.Post("/auth/logout", h.Logout)
	r.Post("/auth/logout-all", h.LogoutAll)
	r.Get("/auth/sso/client", h.SSOClientInfo)
	r.Post("/auth/sso/authorize", h.SSOAuthorize)
	r.Post("/auth/sso/token", h.SSOToken)
	r.Get("/auth/oauth/{provider}", h.OAuthStart)
	r.Get("/auth/{provider}/callback", h.OAuthCallback)
	r.Get("/auth/avatars/{name}", h.ServeAvatar)

	r.Group(func(r chi.Router) {
		r.Use(mw.RequiredJWTAuth(app.DB))
		r.Post("/auth/profile", h.ProfileUpdate)
		r.Post("/auth/avatar", h.UploadAvatar)
		r.Post("/auth/set-password", h.SetPassword)
		r.Get("/auth/identities", h.ListIdentities)
		r.Post("/auth/identities/link", h.LinkIdentity)
		r.Delete("/auth/identities/{provider}", h.UnlinkIdentity)
	})

	// Setup API routes - 使用条件认证中间件，根据路径决定是否需要认证
	r.Group(func(r chi.Router) {
		// JWT 认证中间件
		r.Use(mw.ConditionalJWTAuth(app.DB))

		baseURL := ""
		serverOptions := oapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  mw.HandleBadRequestError(app),
			ResponseErrorHandlerFunc: mw.HandleResponseErrorWithLog(app),
		}
		strictHandler := oapi.NewStrictHandlerWithOptions(
			h,
			[]oapi.StrictMiddlewareFunc{},
			serverOptions,
		)
		oapi.HandlerFromMuxWithBaseURL(strictHandler, r, baseURL)
	})

	return &Server{
		router:     r,
		serverAddr: fmt.Sprintf("%s:%s", app.Server.Host, app.Server.Port),
	}
}

func routeModeFilter(mode Mode) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			isHealth := path == "/health" || path == "/health/"
			isAuth := path == "/auth" || strings.HasPrefix(path, "/auth/")

			if mode == ModeAuth && !isAuth && !isHealth {
				http.NotFound(w, r)
				return
			}
			if mode == ModeQuiz && isAuth {
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
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
