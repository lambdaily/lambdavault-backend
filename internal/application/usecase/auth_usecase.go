package usecase

import (
	"context"

	"github.com/lambdavault/api/internal/application/dto"
	"github.com/lambdavault/api/internal/domain/entity"
	domainErrors "github.com/lambdavault/api/internal/domain/errors"
	"github.com/lambdavault/api/internal/domain/repository"
	"github.com/lambdavault/api/internal/infrastructure/security"
)

type AuthUseCase interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
}

type authUseCase struct {
	userRepo   repository.UserRepository
	groupRepo  repository.PasswordGroupRepository
	hasher     security.Hasher
	jwtService security.JWTService
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	groupRepo repository.PasswordGroupRepository,
	hasher security.Hasher,
	jwtService security.JWTService,
) AuthUseCase {
	return &authUseCase{
		userRepo:   userRepo,
		groupRepo:  groupRepo,
		hasher:     hasher,
		jwtService: jwtService,
	}
}

func (uc *authUseCase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainErrors.ErrUserAlreadyExists
	}

	hash, salt, err := uc.hasher.Hash(req.MasterPassword)
	if err != nil {
		return nil, err
	}

	user := entity.NewUser(req.Email, hash, salt)

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Activate any pending group invitations addressed to this email so the
	// new user immediately sees the groups they were invited to.
	if uc.groupRepo != nil {
		if _, err := uc.groupRepo.ActivatePendingForEmail(ctx, user.Email, user.ID); err != nil {
			return nil, err
		}
	}

	token, err := uc.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	}, nil
}

func (uc *authUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrUserNotFound) {
			return nil, domainErrors.ErrInvalidCredentials
		}
		return nil, err
	}

	valid, err := uc.hasher.Verify(req.MasterPassword, user.MasterPasswordHash, user.Salt)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, domainErrors.ErrInvalidCredentials
	}

	token, err := uc.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	}, nil
}
