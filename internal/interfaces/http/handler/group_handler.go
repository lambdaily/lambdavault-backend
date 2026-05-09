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

// GroupHandler exposes the HTTP endpoints to manage password groups, their
// members and the passwords stored inside each group.
type GroupHandler struct {
	groupUseCase    usecase.GroupUseCase
	passwordUseCase usecase.PasswordUseCase
	validator       *validator.Validator
}

func NewGroupHandler(groupUseCase usecase.GroupUseCase, passwordUseCase usecase.PasswordUseCase, v *validator.Validator) *GroupHandler {
	return &GroupHandler{
		groupUseCase:    groupUseCase,
		passwordUseCase: passwordUseCase,
		validator:       v,
	}
}

// --- Group lifecycle ---

func (h *GroupHandler) Create(c *gin.Context) {
	userID, email, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", h.validator.FormatErrors(err)...)
		return
	}

	result, err := h.groupUseCase.Create(c.Request.Context(), userID, email, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, "group created successfully", result)
}

func (h *GroupHandler) Get(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	result, err := h.groupUseCase.Get(c.Request.Context(), userID, groupID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *GroupHandler) List(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	result, err := h.groupUseCase.List(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *GroupHandler) Update(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", h.validator.FormatErrors(err)...)
		return
	}

	result, err := h.groupUseCase.Update(c.Request.Context(), userID, groupID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *GroupHandler) Delete(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	if err := h.groupUseCase.Delete(c.Request.Context(), userID, groupID); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}

// --- Member management ---

func (h *GroupHandler) AddMembers(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}

	var req dto.AddMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", h.validator.FormatErrors(err)...)
		return
	}

	result, err := h.groupUseCase.AddMembers(c.Request.Context(), userID, groupID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, "members invited successfully", result)
}

func (h *GroupHandler) ListMembers(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	result, err := h.groupUseCase.ListMembers(c.Request.Context(), userID, groupID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *GroupHandler) UpdateMemberRole(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		response.BadRequest(c, "invalid member ID")
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", h.validator.FormatErrors(err)...)
		return
	}
	result, err := h.groupUseCase.UpdateMemberRole(c.Request.Context(), userID, groupID, memberID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		response.BadRequest(c, "invalid member ID")
		return
	}
	if err := h.groupUseCase.RemoveMember(c.Request.Context(), userID, groupID, memberID); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *GroupHandler) Leave(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	if err := h.groupUseCase.Leave(c.Request.Context(), userID, groupID); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}

// --- Group passwords ---

func (h *GroupHandler) CreatePassword(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}

	var req dto.CreateGroupPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", h.validator.FormatErrors(err)...)
		return
	}

	result, err := h.passwordUseCase.CreateInGroup(c.Request.Context(), userID, groupID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, "password created successfully", result)
}

func (h *GroupHandler) AddExistingPassword(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}

	var req dto.AddExistingGroupPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", h.validator.FormatErrors(err)...)
		return
	}

	result, err := h.passwordUseCase.AddExistingToGroup(c.Request.Context(), userID, groupID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, "password added to group successfully", result)
}

func (h *GroupHandler) ListPasswords(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	result, err := h.passwordUseCase.ListGroup(c.Request.Context(), userID, groupID, c.Query("search"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *GroupHandler) GetPassword(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	passwordID, err := uuid.Parse(c.Param("passwordId"))
	if err != nil {
		response.BadRequest(c, "invalid password ID")
		return
	}
	result, err := h.passwordUseCase.GetGroupPassword(c.Request.Context(), userID, groupID, passwordID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *GroupHandler) UpdatePassword(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	passwordID, err := uuid.Parse(c.Param("passwordId"))
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
		response.BadRequest(c, "validation failed", h.validator.FormatErrors(err)...)
		return
	}

	result, err := h.passwordUseCase.UpdateGroupPassword(c.Request.Context(), userID, groupID, passwordID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *GroupHandler) DeletePassword(c *gin.Context) {
	userID, _, err := h.identity(c)
	if err != nil {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	passwordID, err := uuid.Parse(c.Param("passwordId"))
	if err != nil {
		response.BadRequest(c, "invalid password ID")
		return
	}
	if err := h.passwordUseCase.DeleteGroupPassword(c.Request.Context(), userID, groupID, passwordID); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}

// --- helpers ---

func (h *GroupHandler) identity(c *gin.Context) (uuid.UUID, string, error) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, "", domainErrors.ErrMissingAuthHeader
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, "", domainErrors.ErrInvalidToken
	}
	emailValue, _ := c.Get("email")
	email, _ := emailValue.(string)
	return userID, email, nil
}

func (h *GroupHandler) handleError(c *gin.Context, err error) {
	switch {
	case domainErrors.Is(err, domainErrors.ErrGroupNotFound):
		response.NotFound(c, "group not found")
	case domainErrors.Is(err, domainErrors.ErrGroupMemberNotFound):
		response.NotFound(c, "group member not found")
	case domainErrors.Is(err, domainErrors.ErrPasswordNotFound):
		response.NotFound(c, "password not found")
	case domainErrors.Is(err, domainErrors.ErrGroupMemberExists):
		response.Conflict(c, "user is already a member of this group")
	case domainErrors.Is(err, domainErrors.ErrCannotInviteSelf):
		response.BadRequest(c, "you cannot invite yourself")
	case domainErrors.Is(err, domainErrors.ErrInvalidGroupRole):
		response.BadRequest(c, "invalid group role")
	case domainErrors.Is(err, domainErrors.ErrGroupOwnerImmutable):
		response.Forbidden(c, "the group owner cannot be modified or removed")
	case domainErrors.Is(err, domainErrors.ErrInsufficientGroupRole):
		response.Forbidden(c, "you do not have permission to perform this action")
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
