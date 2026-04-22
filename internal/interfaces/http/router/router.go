package router

import (
	"github.com/gin-gonic/gin"

	"github.com/lambdavault/api/internal/application/usecase"
	"github.com/lambdavault/api/internal/domain/repository"
	"github.com/lambdavault/api/internal/infrastructure/config"
	"github.com/lambdavault/api/internal/infrastructure/security"
	"github.com/lambdavault/api/internal/interfaces/http/handler"
	"github.com/lambdavault/api/internal/interfaces/http/middleware"
	"github.com/lambdavault/api/internal/interfaces/http/response"
	"github.com/lambdavault/api/pkg/validator"
)

type Router struct {
	engine         *gin.Engine
	config         *config.Config
	userRepo       repository.UserRepository
	passwordRepo   repository.PasswordRepository
	jwtService     security.JWTService
	hasher         security.Hasher
	encryptor      security.Encryptor
	authMiddleware *middleware.AuthMiddleware
}

func New(
	cfg *config.Config,
	userRepo repository.UserRepository,
	passwordRepo repository.PasswordRepository,
	jwtService security.JWTService,
	hasher security.Hasher,
	encryptor security.Encryptor,
) *Router {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	if cfg.IsDevelopment() {
		engine.Use(gin.Logger())
	}

	return &Router{
		engine:         engine,
		config:         cfg,
		userRepo:       userRepo,
		passwordRepo:   passwordRepo,
		jwtService:     jwtService,
		hasher:         hasher,
		encryptor:      encryptor,
		authMiddleware: middleware.NewAuthMiddleware(jwtService),
	}
}

func (r *Router) Setup() {
	r.setupHealthRoutes()
	r.setupAuthRoutes()
	r.setupProtectedRoutes()
}

func (r *Router) setupHealthRoutes() {
	r.engine.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{
			"status":  "healthy",
			"service": r.config.App.Name,
		})
	})

	r.engine.GET("/ready", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ready"})
	})
}

func (r *Router) setupAuthRoutes() {
	authUseCase := usecase.NewAuthUseCase(r.userRepo, r.hasher, r.jwtService)
	authHandler := handler.NewAuthHandler(authUseCase, validator.New())

	auth := r.engine.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
}

func (r *Router) setupProtectedRoutes() {
	authUseCase := usecase.NewAuthUseCase(r.userRepo, r.hasher, r.jwtService)
	authHandler := handler.NewAuthHandler(authUseCase, validator.New())

	passwordUseCase := usecase.NewPasswordUseCase(r.passwordRepo, r.encryptor)
	passwordHandler := handler.NewPasswordHandler(passwordUseCase, validator.New())

	generatorUseCase := usecase.NewGeneratorUseCase()
	generatorHandler := handler.NewGeneratorHandler(generatorUseCase)

	api := r.engine.Group("/api/v1")
	api.Use(r.authMiddleware.RequireAuth())
	{
		api.GET("/me", authHandler.GetCurrentUser)
		api.GET("/generate-password", generatorHandler.Generate)

		passwords := api.Group("/passwords")
		{
			passwords.POST("", passwordHandler.Create)
			passwords.GET("", passwordHandler.List)
			passwords.GET("/:id", passwordHandler.GetByID)
			passwords.PUT("/:id", passwordHandler.Update)
			passwords.DELETE("/:id", passwordHandler.Delete)
		}
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) Run() error {
	return r.engine.Run(":" + r.config.App.Port)
}
