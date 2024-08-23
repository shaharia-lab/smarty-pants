package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	logger "github.com/chi-middleware/logrus-logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/embedding"
	"github.com/shaharia-lab/smarty-pants/backend/internal/search"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	uuidPath       = "/{uuid}"
	activatePath   = "/activate"
	deactivatePath = "/deactivate"

	invalidUUIDMsg = "Invalid UUID"
)

type API struct {
	config       Config
	router       *chi.Mux
	port         int
	logger       *logrus.Logger
	storage      storage.Storage
	searchSystem search.System
	server       *http.Server
	userManager  *auth.UserManager
	jwtManager   *auth.JWTManager
	aclManager   auth.ACLManager
}

type Config struct {
	Port              int
	ServerReadTimeout int
	WriteTimeout      int
	IdleTimeout       int
}

func NewAPI(logger *logrus.Logger, storage storage.Storage, searchSystem search.System, config Config, userManager *auth.UserManager, jwtManager *auth.JWTManager, aclManager auth.ACLManager) *API {
	api := &API{
		config:       config,
		router:       chi.NewRouter(),
		port:         config.Port,
		logger:       logger,
		storage:      storage,
		searchSystem: searchSystem,
		userManager:  userManager,
		jwtManager:   jwtManager,
		aclManager:   aclManager,
	}
	api.setupMiddleware()
	api.setupRoutes()
	return api
}

func (a *API) setupMiddleware() {
	a.logger.WithField(
		"middlewares",
		[]string{"RequestID", "RealIP", "Logger", "Recoverer", "DetailedRequestLogging"},
	).Debug("setting up middlewares")

	a.router.Use(middleware.RequestID)
	a.router.Use(middleware.RealIP)
	a.router.Use(logger.Logger("router", a.logger))
	a.router.Use(a.enhancedRecoverer)
	a.router.Use(a.detailedRequestLogging)

	a.logger.WithField("timeout", 60*time.Second).Debug("setting up timeout middleware")
	a.router.Use(middleware.Timeout(60 * time.Second))

	cOpts := cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	a.logger.WithFields(logrus.Fields{
		"allowed_origins": cOpts.AllowedOrigins,
		"allowed_methods": cOpts.AllowedMethods,
		"allowed_headers": cOpts.AllowedHeaders,
		"credentials":     cOpts.AllowCredentials,
		"max_age":         cOpts.MaxAge,
	}).Debug("setting up CORS middleware")

	a.router.Use(cors.Handler(cOpts))
}

func (a *API) setupRoutes() {
	type gResponse struct {
		Message string `json:"message"`
	}

	a.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		util.SendSuccessResponse(w, http.StatusOK, gResponse{Message: "Smart assistant!"}, a.logger, nil)
	})

	a.router.Route("/system", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			util.SendSuccessResponse(w, http.StatusOK, gResponse{Message: "Pong"}, a.logger, nil)
		})

		r.Route("/probes", func(r chi.Router) {
			r.Get("/liveness", func(w http.ResponseWriter, r *http.Request) {
				util.SendSuccessResponse(w, http.StatusOK, gResponse{Message: "I am alive"}, a.logger, nil)
			})

			r.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
				util.SendSuccessResponse(w, http.StatusOK, gResponse{Message: "I am ready"}, a.logger, nil)
			})
		})
	})

	a.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(a.jwtManager.AuthMiddleware())

			r.Route("/analytics", func(r chi.Router) {
				r.Get("/overview", getAnalyticsOverview(a.storage, a.logger, a.aclManager))
			})

			r.Route("/datasource", func(r chi.Router) {
				r.Post("/", addDatasourceHandler(a.storage, a.logger))
				r.Route(uuidPath, func(r chi.Router) {
					r.Delete("/", deleteDatasourceHandler(a.storage, a.logger))
					r.Get("/", getDatasourceHandler(a.storage, a.logger))
					r.Get("/validate", validateDatasourceHandler(a.storage, a.logger))
					r.Put("/", updateDatasourceHandler(a.storage, a.logger))
					r.Put(activatePath, setActiveDatasourceHandler(a.storage, a.logger))
					r.Put(deactivatePath, setDisableDatasourceHandler(a.storage, a.logger))
				})
				r.Get("/", getDatasourcesHandler(a.storage, a.logger))
			})

			r.Route("/document", func(r chi.Router) {
				r.Get(uuidPath, getDocumentHandler(a.storage, a.logger))
				r.Get("/", getDocumentsHandler(a.storage, a.logger))
			})

			r.Route("/embedding-provider", func(r chi.Router) {
				r.Post("/", addEmbeddingProviderHandler(a.storage, a.logger))
				r.Route(uuidPath, func(r chi.Router) {
					r.Delete("/", deleteEmbeddingProviderHandler(a.storage, a.logger))
					r.Get("/", getEmbeddingProviderHandler(a.storage, a.logger))
					r.Put("/", updateEmbeddingProviderHandler(a.storage, a.logger))
					r.Put(activatePath, setActiveEmbeddingProviderHandler(a.storage, a.logger))
					r.Put(deactivatePath, setDisableEmbeddingProviderHandler(a.storage, a.logger))
				})
				r.Get("/", getEmbeddingProvidersHandler(a.storage, a.logger))
			})

			r.Route("/interactions", func(r chi.Router) {
				r.Post("/", createInteractionHandler(a.storage, a.logger))
				r.Get("/", getInteractionsHandler(a.logger))
				r.Route(uuidPath, func(r chi.Router) {
					r.Get("/", getInteractionHandler(a.logger))
					r.Post("/message", sendMessageHandler(a.searchSystem, a.storage, a.logger))
				})
			})

			r.Route("/llm-provider", func(r chi.Router) {
				r.Post("/", addLLMProviderHandler(a.storage, a.logger))
				r.Route(uuidPath, func(r chi.Router) {
					r.Delete("/", deleteLLMProviderHandler(a.storage, a.logger))
					r.Get("/", getLLMProviderHandler(a.storage, a.logger))
					r.Put("/", updateLLMProviderHandler(a.storage, a.logger))
					r.Put(activatePath, setActiveLLMProviderHandler(a.storage, a.logger))
					r.Put(deactivatePath, setDisableLLMProviderHandler(a.storage, a.logger))
				})
				r.Get("/", getLLMProvidersHandler(a.storage, a.logger))
			})

			r.Route("/search", func(r chi.Router) {
				r.Post("/", addSearchHandler(a.searchSystem, a.logger))
			})

			r.Route("/settings", func(r chi.Router) {
				r.Get("/", getSettingsHandler(a.storage, a.logger))
				r.Put("/", updateSettingsHandler(a.storage, a.logger))
			})
		})
	})

	a.userManager.RegisterRoutes(a.router)
}

// Start starts the API server
func (a *API) Start(ctx context.Context) error {
	a.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.config.Port),
		Handler:      a.router,
		ReadTimeout:  time.Duration(a.config.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(a.config.IdleTimeout) * time.Second,
	}

	go a.startMemoryUsageLogging(ctx)
	go a.startDependencyHealthLogging(ctx)

	a.logger.WithField("port", a.config.Port).Info("Starting API server")
	err := a.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (a *API) StartWithHealthCheck(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.Start(gCtx)
	})

	g.Go(func() error {
		healthServer := &http.Server{
			Addr: ":8081",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		}

		go func() {
			<-gCtx.Done()
			_ = healthServer.Shutdown(context.Background())
		}()

		return healthServer.ListenAndServe()
	})

	return g.Wait()
}

func (a *API) detailedRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(rww, r)

		duration := time.Since(start)

		a.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"duration":   duration,
			"status":     rww.Status(),
			"size":       rww.BytesWritten(),
			"user_agent": r.UserAgent(),
		}).Info("Request completed")
	})
}

func (a *API) enhancedRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				a.logger.WithFields(logrus.Fields{
					"panic": rvr,
					"stack": string(debug.Stack()),
				}).Error("Panic recovered")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *API) startMemoryUsageLogging(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			a.logger.WithFields(logrus.Fields{
				"alloc_mb":       m.Alloc / 1024 / 1024,
				"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
				"sys_mb":         m.Sys / 1024 / 1024,
				"num_gc":         m.NumGC,
			}).Debug("Memory usage stats")
		case <-ctx.Done():
			return
		}
	}
}

func (a *API) startDependencyHealthLogging(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.storage.HealthCheck(); err != nil {
				a.logger.WithError(err).Error("Storage health check failed")
			} else {
				a.logger.Debug("Storafge health check passed")
			}

			emProvider, err := embedding.InitializeEmbeddingProvider(ctx, a.storage, a.logger)
			if err != nil || emProvider == nil {
				return
			}

			if err := emProvider.HealthCheck(); err != nil {
				a.logger.WithError(err).Error("Embedding provider health check failed")
			} else {
				a.logger.Debug("Embedding provider health check passed")
			}

			if err := a.searchSystem.HealthCheck(); err != nil {
				a.logger.WithError(err).Error("Search system health check failed")
			} else {
				a.logger.Debug("Search system health check passed")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *API) Shutdown(ctx context.Context) error {
	a.logger.Info("Initiating API server shutdown")

	if a.server == nil {
		a.logger.Warn("API server is not running")
		return nil
	}

	err := a.server.Shutdown(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			a.logger.Warn("API server shutdown timed out")
		} else {
			a.logger.WithError(err).Error("Error during API server shutdown")
		}
		return err
	}

	a.logger.Info("API server shutdown completed successfully")
	return nil
}
