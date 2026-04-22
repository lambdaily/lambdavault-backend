package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/lambdavault/api/internal/application/dto"
	"github.com/lambdavault/api/internal/application/usecase"
	domainErrors "github.com/lambdavault/api/internal/domain/errors"
	"github.com/lambdavault/api/internal/interfaces/http/response"
	"github.com/lambdavault/api/pkg/validator"
)

type PasswordHandler struct {
	passwordUseCase usecase.PasswordUseCase
	validator       *validator.Validator
}

func NewPasswordHandler(passwordUseCase usecase.PasswordUseCase, validator *validator.Validator) *PasswordHandler {
	return &PasswordHandler{
		passwordUseCase: passwordUseCase,
		validator:       validator,
	}
}

func (h *PasswordHandler) Create(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var req dto.CreatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.validator.FormatErrors(err)
		response.BadRequest(c, "validation failed", errors...)
		return
	}

	result, err := h.passwordUseCase.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, "password created successfully", result)
}

func (h *PasswordHandler) GetByID(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	passwordID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid password ID")
		return
	}

	result, err := h.passwordUseCase.GetByID(c.Request.Context(), userID, passwordID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *PasswordHandler) List(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	searchQuery := c.Query("search")
	var result *dto.PasswordListResponse

	if searchQuery != "" {
		result, err = h.passwordUseCase.Search(c.Request.Context(), userID, searchQuery)
	} else {
		result, err = h.passwordUseCase.List(c.Request.Context(), userID)
	}

	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *PasswordHandler) Update(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	passwordID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid password ID")
		return
	}

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.validator.FormatErrors(err)
		response.BadRequest(c, "validation failed", errors...)
		return
	}

	result, err := h.passwordUseCase.Update(c.Request.Context(), userID, passwordID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *PasswordHandler) Delete(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	passwordID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid password ID")
		return
	}

	if err := h.passwordUseCase.Delete(c.Request.Context(), userID, passwordID); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *PasswordHandler) getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, domainErrors.ErrMissingAuthHeader
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, domainErrors.ErrInvalidToken
	}

	return userID, nil
}

func (h *PasswordHandler) handleError(c *gin.Context, err error) {
	switch {
	case domainErrors.Is(err, domainErrors.ErrPasswordNotFound):
		response.NotFound(c, "password not found")
	case domainErrors.Is(err, domainErrors.ErrAccessDenied):
		response.Forbidden(c, "access denied")
	case domainErrors.Is(err, domainErrors.ErrEncryptionFailed):
		response.InternalServerError(c, "encryption failed")
	case domainErrors.Is(err, domainErrors.ErrDecryptionFailed):
		response.InternalServerError(c, "decryption failed")
	default:
		response.InternalServerError(c, "an error occurred")
	}
}
