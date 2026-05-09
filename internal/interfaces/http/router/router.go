package router

import (
	"github.com/gin-gonic/gin"

	"github.com/lambdavault/api/internal/application/usecase"
	"github.com/lambdavault/api/internal/domain/repository"
	"github.com/lambdavault/api/internal/infrastructure/config"
	"github.com/lambdavault/api/internal/infrastructure/persistence/database"
	"github.com/lambdavault/api/internal/infrastructure/security"
	"github.com/lambdavault/api/internal/interfaces/http/handler"
	"github.com/lambdavault/api/internal/interfaces/http/middleware"
	"github.com/lambdavault/api/internal/interfaces/http/response"
	"github.com/lambdavault/api/pkg/validator"
)

type Router struct {
	engine         *gin.Engine
	config         *config.Config
	db             *database.Database
	userRepo       repository.UserRepository
	passwordRepo   repository.PasswordRepository
	groupRepo      repository.PasswordGroupRepository
	jwtService     security.JWTService
	hasher         security.Hasher
	encryptor      security.Encryptor
	authMiddleware *middleware.AuthMiddleware
}

func New(
	cfg *config.Config,
	userRepo repository.UserRepository,
	passwordRepo repository.PasswordRepository,
	groupRepo repository.PasswordGroupRepository,
	jwtService security.JWTService,
	hasher security.Hasher,
	encryptor security.Encryptor,
	db *database.Database,
) *Router {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.SecurityHeaders(cfg.IsProduction()))
	engine.Use(middleware.CORS(cfg.CORS.Origins, cfg.IsProduction()))
	engine.Use(middleware.RateLimit(cfg.RateLimit.RPS, cfg.RateLimit.Burst))

	if cfg.IsDevelopment() {
		engine.Use(gin.Logger())
	}

	return &Router{
		engine:         engine,
		config:         cfg,
		db:             db,
		userRepo:       userRepo,
		passwordRepo:   passwordRepo,
		groupRepo:      groupRepo,
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
		if r.db != nil {
			if err := r.db.Ping(); err != nil {
				c.AbortWithStatusJSON(503, gin.H{"status": "not_ready", "error": err.Error()})
				return
			}
		}
		response.OK(c, gin.H{"status": "ready"})
	})
}

func (r *Router) setupAuthRoutes() {
	authUseCase := usecase.NewAuthUseCase(r.userRepo, r.groupRepo, r.hasher, r.jwtService)
	authHandler := handler.NewAuthHandler(authUseCase, validator.New())

	auth := r.engine.Group("/api/v1/auth")
	auth.Use(middleware.RateLimit(r.config.RateLimit.AuthRPS, r.config.RateLimit.AuthBurst))
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
}

func (r *Router) setupProtectedRoutes() {
	authUseCase := usecase.NewAuthUseCase(r.userRepo, r.groupRepo, r.hasher, r.jwtService)
	authHandler := handler.NewAuthHandler(authUseCase, validator.New())

	passwordUseCase := usecase.NewPasswordUseCase(r.passwordRepo, r.groupRepo, r.encryptor)
	passwordHandler := handler.NewPasswordHandler(passwordUseCase, validator.New())

	groupUseCase := usecase.NewGroupUseCase(r.groupRepo, r.userRepo)
	groupHandler := handler.NewGroupHandler(groupUseCase, passwordUseCase, validator.New())

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

		groups := api.Group("/groups")
		{
			groups.POST("", groupHandler.Create)
			groups.GET("", groupHandler.List)
			groups.GET("/:id", groupHandler.Get)
			groups.PUT("/:id", groupHandler.Update)
			groups.DELETE("/:id", groupHandler.Delete)
			groups.POST("/:id/leave", groupHandler.Leave)

			groups.GET("/:id/members", groupHandler.ListMembers)
			groups.POST("/:id/members", groupHandler.AddMembers)
			groups.PUT("/:id/members/:memberId", groupHandler.UpdateMemberRole)
			groups.DELETE("/:id/members/:memberId", groupHandler.RemoveMember)

			groups.GET("/:id/passwords", groupHandler.ListPasswords)
			groups.POST("/:id/passwords", groupHandler.CreatePassword)
			groups.GET("/:id/passwords/:passwordId", groupHandler.GetPassword)
			groups.PUT("/:id/passwords/:passwordId", groupHandler.UpdatePassword)
			groups.DELETE("/:id/passwords/:passwordId", groupHandler.DeletePassword)
		}
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) Run() error {
	return r.engine.Run(":" + r.config.App.Port)
}
