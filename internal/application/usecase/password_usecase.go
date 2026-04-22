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
	Create(ctx context.Context, userID uuid.UUID, req dto.CreatePasswordRequest) (*dto.PasswordResponse, error)
	GetByID(ctx context.Context, userID, passwordID uuid.UUID) (*dto.PasswordWithSecretResponse, error)
	List(ctx context.Context, userID uuid.UUID) (*dto.PasswordListResponse, error)
	Search(ctx context.Context, userID uuid.UUID, query string) (*dto.PasswordListResponse, error)
	Update(ctx context.Context, userID, passwordID uuid.UUID, req dto.UpdatePasswordRequest) (*dto.PasswordResponse, error)
	Delete(ctx context.Context, userID, passwordID uuid.UUID) error
}

type passwordUseCase struct {
	passwordRepo repository.PasswordRepository
	encryptor    security.Encryptor
}

func NewPasswordUseCase(
	passwordRepo repository.PasswordRepository,
	encryptor security.Encryptor,
) PasswordUseCase {
	return &passwordUseCase{
		passwordRepo: passwordRepo,
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

	responses := make([]dto.PasswordResponse, len(passwords))
	for i, p := range passwords {
		responses[i] = *uc.toResponse(p)
	}

	return &dto.PasswordListResponse{
		Passwords: responses,
		Total:     len(responses),
	}, nil
}

func (uc *passwordUseCase) Search(ctx context.Context, userID uuid.UUID, query string) (*dto.PasswordListResponse, error) {
	passwords, err := uc.passwordRepo.SearchByUserID(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PasswordResponse, len(passwords))
	for i, p := range passwords {
		responses[i] = *uc.toResponse(p)
	}

	return &dto.PasswordListResponse{
		Passwords: responses,
		Total:     len(responses),
	}, nil
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

func (uc *passwordUseCase) toResponse(p *entity.Password) *dto.PasswordResponse {
	return &dto.PasswordResponse{
		ID:        p.ID,
		SiteName:  p.SiteName,
		SiteURL:   p.SiteURL,
		Username:  p.Username,
		Notes:     p.Notes,
		Category:  p.Category,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
