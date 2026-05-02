package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/lambdavault/api/internal/application/dto"
	"github.com/lambdavault/api/internal/domain/entity"
	domainErrors "github.com/lambdavault/api/internal/domain/errors"
	"github.com/lambdavault/api/internal/domain/repository"
	"github.com/lambdavault/api/internal/infrastructure/security"
)

type PasswordUseCase interface {
	// Personal vault operations.
	Create(ctx context.Context, userID uuid.UUID, req dto.CreatePasswordRequest) (*dto.PasswordResponse, error)
	GetByID(ctx context.Context, userID, passwordID uuid.UUID) (*dto.PasswordWithSecretResponse, error)
	List(ctx context.Context, userID uuid.UUID) (*dto.PasswordListResponse, error)
	Search(ctx context.Context, userID uuid.UUID, query string) (*dto.PasswordListResponse, error)
	Update(ctx context.Context, userID, passwordID uuid.UUID, req dto.UpdatePasswordRequest) (*dto.PasswordResponse, error)
	Delete(ctx context.Context, userID, passwordID uuid.UUID) error

	// Group vault operations. All of them resolve the caller's role inside the
	// group and enforce the permission flags configured by the admin.
	CreateInGroup(ctx context.Context, userID, groupID uuid.UUID, req dto.CreateGroupPasswordRequest) (*dto.PasswordResponse, error)
	GetGroupPassword(ctx context.Context, userID, groupID, passwordID uuid.UUID) (*dto.PasswordWithSecretResponse, error)
	ListGroup(ctx context.Context, userID, groupID uuid.UUID, query string) (*dto.PasswordListResponse, error)
	UpdateGroupPassword(ctx context.Context, userID, groupID, passwordID uuid.UUID, req dto.UpdatePasswordRequest) (*dto.PasswordResponse, error)
	DeleteGroupPassword(ctx context.Context, userID, groupID, passwordID uuid.UUID) error
}

type passwordUseCase struct {
	passwordRepo repository.PasswordRepository
	groupRepo    repository.PasswordGroupRepository
	encryptor    security.Encryptor
}

func NewPasswordUseCase(
	passwordRepo repository.PasswordRepository,
	groupRepo repository.PasswordGroupRepository,
	encryptor security.Encryptor,
) PasswordUseCase {
	return &passwordUseCase{
		passwordRepo: passwordRepo,
		groupRepo:    groupRepo,
		encryptor:    encryptor,
	}
}

func (uc *passwordUseCase) Create(ctx context.Context, userID uuid.UUID, req dto.CreatePasswordRequest) (*dto.PasswordResponse, error) {
	encryptedPassword, iv, err := uc.encryptor.Encrypt(req.Password)
	if err != nil {
		return nil, domainErrors.ErrEncryptionFailed
	}

	password := entity.NewPassword(
		userID,
		req.SiteName,
		req.SiteURL,
		req.Username,
		encryptedPassword,
		iv,
		req.Notes,
		req.Category,
	)

	if err := uc.passwordRepo.Create(ctx, password); err != nil {
		return nil, err
	}

	return uc.toResponse(password), nil
}

func (uc *passwordUseCase) GetByID(ctx context.Context, userID, passwordID uuid.UUID) (*dto.PasswordWithSecretResponse, error) {
	password, err := uc.passwordRepo.FindByIDAndUserID(ctx, passwordID, userID)
	if err != nil {
		return nil, err
	}

	decryptedPassword, err := uc.encryptor.Decrypt(password.EncryptedPassword, password.IV)
	if err != nil {
		return nil, domainErrors.ErrDecryptionFailed
	}

	return &dto.PasswordWithSecretResponse{
		PasswordResponse: *uc.toResponse(password),
		Password:         decryptedPassword,
	}, nil
}

func (uc *passwordUseCase) List(ctx context.Context, userID uuid.UUID) (*dto.PasswordListResponse, error) {
	passwords, err := uc.passwordRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.toListResponse(passwords), nil
}

func (uc *passwordUseCase) Search(ctx context.Context, userID uuid.UUID, query string) (*dto.PasswordListResponse, error) {
	passwords, err := uc.passwordRepo.SearchByUserID(ctx, userID, query)
	if err != nil {
		return nil, err
	}
	return uc.toListResponse(passwords), nil
}

func (uc *passwordUseCase) Update(ctx context.Context, userID, passwordID uuid.UUID, req dto.UpdatePasswordRequest) (*dto.PasswordResponse, error) {
	password, err := uc.passwordRepo.FindByIDAndUserID(ctx, passwordID, userID)
	if err != nil {
		return nil, err
	}

	encryptedPassword, iv, err := uc.encryptor.Encrypt(req.Password)
	if err != nil {
		return nil, domainErrors.ErrEncryptionFailed
	}

	password.Update(req.SiteName, req.SiteURL, req.Username, encryptedPassword, iv, req.Notes, req.Category)

	if err := uc.passwordRepo.Update(ctx, password); err != nil {
		return nil, err
	}

	return uc.toResponse(password), nil
}

func (uc *passwordUseCase) Delete(ctx context.Context, userID, passwordID uuid.UUID) error {
	_, err := uc.passwordRepo.FindByIDAndUserID(ctx, passwordID, userID)
	if err != nil {
		return err
	}

	return uc.passwordRepo.Delete(ctx, passwordID)
}

func (uc *passwordUseCase) CreateInGroup(ctx context.Context, userID, groupID uuid.UUID, req dto.CreateGroupPasswordRequest) (*dto.PasswordResponse, error) {
	group, role, err := uc.requireGroupAccess(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if !group.CanCreatePasswords(role) {
		return nil, domainErrors.ErrInsufficientGroupRole
	}

	encryptedPassword, iv, err := uc.encryptor.Encrypt(req.Password)
	if err != nil {
		return nil, domainErrors.ErrEncryptionFailed
	}

	password := entity.NewGroupPassword(
		userID,
		groupID,
		req.SiteName,
		req.SiteURL,
		req.Username,
		encryptedPassword,
		iv,
		req.Notes,
		req.Category,
	)
	if err := uc.passwordRepo.Create(ctx, password); err != nil {
		return nil, err
	}
	return uc.toResponse(password), nil
}

func (uc *passwordUseCase) GetGroupPassword(ctx context.Context, userID, groupID, passwordID uuid.UUID) (*dto.PasswordWithSecretResponse, error) {
	if _, _, err := uc.requireGroupAccess(ctx, userID, groupID); err != nil {
		return nil, err
	}
	password, err := uc.passwordRepo.FindByID(ctx, passwordID)
	if err != nil {
		return nil, err
	}
	if password.GroupID == nil || *password.GroupID != groupID {
		return nil, domainErrors.ErrPasswordNotFound
	}

	decrypted, err := uc.encryptor.Decrypt(password.EncryptedPassword, password.IV)
	if err != nil {
		return nil, domainErrors.ErrDecryptionFailed
	}
	return &dto.PasswordWithSecretResponse{
		PasswordResponse: *uc.toResponse(password),
		Password:         decrypted,
	}, nil
}

func (uc *passwordUseCase) ListGroup(ctx context.Context, userID, groupID uuid.UUID, query string) (*dto.PasswordListResponse, error) {
	if _, _, err := uc.requireGroupAccess(ctx, userID, groupID); err != nil {
		return nil, err
	}
	var (
		passwords []*entity.Password
		err       error
	)
	if query == "" {
		passwords, err = uc.passwordRepo.FindAllByGroupID(ctx, groupID)
	} else {
		passwords, err = uc.passwordRepo.SearchByGroupID(ctx, groupID, query)
	}
	if err != nil {
		return nil, err
	}
	return uc.toListResponse(passwords), nil
}

func (uc *passwordUseCase) UpdateGroupPassword(ctx context.Context, userID, groupID, passwordID uuid.UUID, req dto.UpdatePasswordRequest) (*dto.PasswordResponse, error) {
	group, role, err := uc.requireGroupAccess(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if !group.CanEditPasswords(role) {
		return nil, domainErrors.ErrInsufficientGroupRole
	}
	password, err := uc.passwordRepo.FindByID(ctx, passwordID)
	if err != nil {
		return nil, err
	}
	if password.GroupID == nil || *password.GroupID != groupID {
		return nil, domainErrors.ErrPasswordNotFound
	}

	encryptedPassword, iv, err := uc.encryptor.Encrypt(req.Password)
	if err != nil {
		return nil, domainErrors.ErrEncryptionFailed
	}
	password.Update(req.SiteName, req.SiteURL, req.Username, encryptedPassword, iv, req.Notes, req.Category)
	if err := uc.passwordRepo.Update(ctx, password); err != nil {
		return nil, err
	}
	return uc.toResponse(password), nil
}

func (uc *passwordUseCase) DeleteGroupPassword(ctx context.Context, userID, groupID, passwordID uuid.UUID) error {
	group, role, err := uc.requireGroupAccess(ctx, userID, groupID)
	if err != nil {
		return err
	}
	if !group.CanDeletePasswords(role) {
		return domainErrors.ErrInsufficientGroupRole
	}
	password, err := uc.passwordRepo.FindByID(ctx, passwordID)
	if err != nil {
		return err
	}
	if password.GroupID == nil || *password.GroupID != groupID {
		return domainErrors.ErrPasswordNotFound
	}
	return uc.passwordRepo.Delete(ctx, passwordID)
}

// requireGroupAccess loads the group and resolves the caller's role,
// returning ErrGroupNotFound when the user has no relationship with the
// group so we don't disclose its existence.
func (uc *passwordUseCase) requireGroupAccess(ctx context.Context, userID, groupID uuid.UUID) (*entity.PasswordGroup, entity.GroupRole, error) {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, "", err
	}
	if group.OwnerID == userID {
		return group, entity.GroupRoleAdmin, nil
	}
	member, err := uc.groupRepo.FindMemberByGroupAndUser(ctx, groupID, userID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrGroupMemberNotFound) {
			return nil, "", domainErrors.ErrGroupNotFound
		}
		return nil, "", err
	}
	if member.Status != entity.GroupMemberStatusActive {
		return nil, "", domainErrors.ErrGroupNotFound
	}
	return group, member.Role, nil
}

func (uc *passwordUseCase) toListResponse(passwords []*entity.Password) *dto.PasswordListResponse {
	responses := make([]dto.PasswordResponse, len(passwords))
	for i, p := range passwords {
		responses[i] = *uc.toResponse(p)
	}
	return &dto.PasswordListResponse{Passwords: responses, Total: len(responses)}
}

func (uc *passwordUseCase) toResponse(p *entity.Password) *dto.PasswordResponse {
	return &dto.PasswordResponse{
		ID:        p.ID,
		GroupID:   p.GroupID,
		SiteName:  p.SiteName,
		SiteURL:   p.SiteURL,
		Username:  p.Username,
		Notes:     p.Notes,
		Category:  p.Category,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
