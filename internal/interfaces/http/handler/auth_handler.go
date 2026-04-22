package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/lambdavault/api/internal/application/dto"
	"github.com/lambdavault/api/internal/application/usecase"
	domainErrors "github.com/lambdavault/api/internal/domain/errors"
	"github.com/lambdavault/api/internal/interfaces/http/response"
	"github.com/lambdavault/api/pkg/validator"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
	validator   *validator.Validator
}

func NewAuthHandler(authUseCase usecase.AuthUseCase, validator *validator.Validator) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		validator:   validator,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.validator.FormatErrors(err)
		response.BadRequest(c, "validation failed", errors...)
		return
	}

	result, err := h.authUseCase.Register(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, "user registered successfully", result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.validator.FormatErrors(err)
		response.BadRequest(c, "validation failed", errors...)
		return
	}

	result, err := h.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	switch {
	case domainErrors.Is(err, domainErrors.ErrUserAlreadyExists):
		response.Conflict(c, "user already exists")
	case domainErrors.Is(err, domainErrors.ErrInvalidCredentials):
		response.Unauthorized(c, "invalid email or password")
	case domainErrors.Is(err, domainErrors.ErrUserNotFound):
		response.NotFound(c, "user not found")
	default:
		response.InternalServerError(c, "an error occurred")
	}
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	email, _ := c.Get("email")

	response.OK(c, gin.H{
		"id":    userID,
		"email": email,
	})
}
