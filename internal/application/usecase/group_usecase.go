package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/lambdavault/api/internal/application/dto"
	"github.com/lambdavault/api/internal/domain/entity"
	domainErrors "github.com/lambdavault/api/internal/domain/errors"
	"github.com/lambdavault/api/internal/domain/repository"
)

// GroupUseCase exposes operations on password groups and their members.
//
// Authorization rules:
//   - The owner is always treated as an admin.
//   - Only admins can update group settings, invite/remove members or change
//     a member's role.
//   - Only the owner can delete the group.
type GroupUseCase interface {
	Create(ctx context.Context, ownerID uuid.UUID, ownerEmail string, req dto.CreateGroupRequest) (*dto.GroupDetailResponse, error)
	Get(ctx context.Context, userID, groupID uuid.UUID) (*dto.GroupDetailResponse, error)
	List(ctx context.Context, userID uuid.UUID) (*dto.GroupListResponse, error)
	Update(ctx context.Context, userID, groupID uuid.UUID, req dto.UpdateGroupRequest) (*dto.GroupResponse, error)
	Delete(ctx context.Context, userID, groupID uuid.UUID) error

	AddMembers(ctx context.Context, userID, groupID uuid.UUID, req dto.AddMembersRequest) (*dto.GroupMembersResponse, error)
	ListMembers(ctx context.Context, userID, groupID uuid.UUID) (*dto.GroupMembersResponse, error)
	UpdateMemberRole(ctx context.Context, userID, groupID, memberID uuid.UUID, req dto.UpdateMemberRoleRequest) (*dto.GroupMemberResponse, error)
	RemoveMember(ctx context.Context, userID, groupID, memberID uuid.UUID) error
	Leave(ctx context.Context, userID, groupID uuid.UUID) error
}

type groupUseCase struct {
	groupRepo repository.PasswordGroupRepository
	userRepo  repository.UserRepository
}

func NewGroupUseCase(groupRepo repository.PasswordGroupRepository, userRepo repository.UserRepository) GroupUseCase {
	return &groupUseCase{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// Create persists a new group, makes the creator an admin and immediately
// dispatches any invitations included in the request.
func (uc *groupUseCase) Create(ctx context.Context, ownerID uuid.UUID, ownerEmail string, req dto.CreateGroupRequest) (*dto.GroupDetailResponse, error) {
	group := entity.NewPasswordGroup(ownerID, req.Name, req.Description, req.AllowMemberCreate, req.AllowMemberEdit, req.AllowMemberDelete)
	if err := uc.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	// Owner gets an explicit membership row with the admin role so member
	// listings include them and role checks have a single code path.
	owner, err := uc.userRepo.FindByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	ownerMember := entity.NewGroupMember(group.ID, ownerID, owner.Email, entity.GroupRoleAdmin, owner)
	if err := uc.groupRepo.AddMember(ctx, ownerMember); err != nil {
		return nil, err
	}

	// Invite extra members if provided. Errors here are non-fatal in the sense
	// that they are returned but the group is already created.
	for _, invite := range req.Members {
		if _, err := uc.inviteOne(ctx, group.ID, ownerID, ownerEmail, invite); err != nil {
			return nil, err
		}
	}

	return uc.buildDetail(ctx, group, ownerID)
}

func (uc *groupUseCase) Get(ctx context.Context, userID, groupID uuid.UUID) (*dto.GroupDetailResponse, error) {
	group, _, err := uc.requireMember(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	return uc.buildDetail(ctx, group, userID)
}

func (uc *groupUseCase) List(ctx context.Context, userID uuid.UUID) (*dto.GroupListResponse, error) {
	groups, err := uc.groupRepo.FindAccessibleByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.GroupResponse, 0, len(groups))
	for _, g := range groups {
		members, err := uc.groupRepo.ListMembers(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		role := uc.callerRole(g, members, userID)
		responses = append(responses, toGroupResponse(g, role, len(members)))
	}

	return &dto.GroupListResponse{Groups: responses, Total: len(responses)}, nil
}

func (uc *groupUseCase) Update(ctx context.Context, userID, groupID uuid.UUID, req dto.UpdateGroupRequest) (*dto.GroupResponse, error) {
	group, role, err := uc.requireMember(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if role != entity.GroupRoleAdmin {
		return nil, domainErrors.ErrInsufficientGroupRole
	}

	group.UpdateSettings(req.Name, req.Description, req.AllowMemberCreate, req.AllowMemberEdit, req.AllowMemberDelete)
	if err := uc.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	members, err := uc.groupRepo.ListMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	resp := toGroupResponse(group, role, len(members))
	return &resp, nil
}

func (uc *groupUseCase) Delete(ctx context.Context, userID, groupID uuid.UUID) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.OwnerID != userID {
		return domainErrors.ErrInsufficientGroupRole
	}
	return uc.groupRepo.Delete(ctx, groupID)
}

func (uc *groupUseCase) AddMembers(ctx context.Context, userID, groupID uuid.UUID, req dto.AddMembersRequest) (*dto.GroupMembersResponse, error) {
	group, role, err := uc.requireMember(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if role != entity.GroupRoleAdmin {
		return nil, domainErrors.ErrInsufficientGroupRole
	}

	owner, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	added := make([]dto.GroupMemberResponse, 0, len(req.Members))
	for _, invite := range req.Members {
		member, err := uc.inviteOne(ctx, group.ID, userID, owner.Email, invite)
		if err != nil {
			return nil, err
		}
		added = append(added, toMemberResponse(member))
	}

	return &dto.GroupMembersResponse{Members: added, Total: len(added)}, nil
}

func (uc *groupUseCase) ListMembers(ctx context.Context, userID, groupID uuid.UUID) (*dto.GroupMembersResponse, error) {
	if _, _, err := uc.requireMember(ctx, userID, groupID); err != nil {
		return nil, err
	}
	members, err := uc.groupRepo.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.GroupMemberResponse, len(members))
	for i, m := range members {
		out[i] = toMemberResponse(m)
	}
	return &dto.GroupMembersResponse{Members: out, Total: len(out)}, nil
}

func (uc *groupUseCase) UpdateMemberRole(ctx context.Context, userID, groupID, memberID uuid.UUID, req dto.UpdateMemberRoleRequest) (*dto.GroupMemberResponse, error) {
	group, role, err := uc.requireMember(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if role != entity.GroupRoleAdmin {
		return nil, domainErrors.ErrInsufficientGroupRole
	}

	newRole := entity.GroupRole(req.Role)
	if !newRole.IsValid() {
		return nil, domainErrors.ErrInvalidGroupRole
	}

	member, err := uc.groupRepo.FindMemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member.GroupID != group.ID {
		return nil, domainErrors.ErrGroupMemberNotFound
	}
	if member.UserID != nil && *member.UserID == group.OwnerID {
		return nil, domainErrors.ErrGroupOwnerImmutable
	}

	member.ChangeRole(newRole)
	if err := uc.groupRepo.UpdateMember(ctx, member); err != nil {
		return nil, err
	}
	resp := toMemberResponse(member)
	return &resp, nil
}

func (uc *groupUseCase) RemoveMember(ctx context.Context, userID, groupID, memberID uuid.UUID) error {
	group, role, err := uc.requireMember(ctx, userID, groupID)
	if err != nil {
		return err
	}
	if role != entity.GroupRoleAdmin {
		return domainErrors.ErrInsufficientGroupRole
	}

	member, err := uc.groupRepo.FindMemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member.GroupID != group.ID {
		return domainErrors.ErrGroupMemberNotFound
	}
	if member.UserID != nil && *member.UserID == group.OwnerID {
		return domainErrors.ErrGroupOwnerImmutable
	}
	return uc.groupRepo.RemoveMember(ctx, member.ID)
}

func (uc *groupUseCase) Leave(ctx context.Context, userID, groupID uuid.UUID) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.OwnerID == userID {
		// The owner has to delete the group instead of leaving it.
		return domainErrors.ErrGroupOwnerImmutable
	}
	member, err := uc.groupRepo.FindMemberByGroupAndUser(ctx, groupID, userID)
	if err != nil {
		return err
	}
	return uc.groupRepo.RemoveMember(ctx, member.ID)
}

// inviteOne adds a single member to a group. If the email already corresponds
// to a registered user the membership is created in the active state;
// otherwise it is stored as a pending invitation that will be activated when
// that user signs up.
func (uc *groupUseCase) inviteOne(ctx context.Context, groupID, inviterID uuid.UUID, inviterEmail string, invite dto.GroupInviteEntry) (*entity.GroupMember, error) {
	role := entity.GroupRole(invite.Role)
	if !role.IsValid() {
		return nil, domainErrors.ErrInvalidGroupRole
	}

	email := strings.ToLower(strings.TrimSpace(invite.Email))
	if email == strings.ToLower(strings.TrimSpace(inviterEmail)) {
		return nil, domainErrors.ErrCannotInviteSelf
	}

	if existing, err := uc.groupRepo.FindMemberByGroupAndEmail(ctx, groupID, email); err == nil {
		_ = existing
		return nil, domainErrors.ErrGroupMemberExists
	} else if !domainErrors.Is(err, domainErrors.ErrGroupMemberNotFound) {
		return nil, err
	}

	var user *entity.User
	if found, err := uc.userRepo.FindByEmail(ctx, email); err == nil {
		user = found
	} else if !domainErrors.Is(err, domainErrors.ErrUserNotFound) {
		return nil, err
	}

	member := entity.NewGroupMember(groupID, inviterID, email, role, user)
	if err := uc.groupRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}
	return member, nil
}

// requireMember loads the group and resolves the caller's effective role.
// Returns ErrGroupNotFound when the user has no relationship with the group
// (no information leakage about its existence).
func (uc *groupUseCase) requireMember(ctx context.Context, userID, groupID uuid.UUID) (*entity.PasswordGroup, entity.GroupRole, error) {
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

func (uc *groupUseCase) callerRole(group *entity.PasswordGroup, members []*entity.GroupMember, userID uuid.UUID) entity.GroupRole {
	if group.OwnerID == userID {
		return entity.GroupRoleAdmin
	}
	for _, m := range members {
		if m.UserID != nil && *m.UserID == userID && m.Status == entity.GroupMemberStatusActive {
			return m.Role
		}
	}
	return ""
}

func (uc *groupUseCase) buildDetail(ctx context.Context, group *entity.PasswordGroup, userID uuid.UUID) (*dto.GroupDetailResponse, error) {
	members, err := uc.groupRepo.ListMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	role := uc.callerRole(group, members, userID)
	memberDTOs := make([]dto.GroupMemberResponse, len(members))
	for i, m := range members {
		memberDTOs[i] = toMemberResponse(m)
	}
	return &dto.GroupDetailResponse{
		Group:   toGroupResponse(group, role, len(members)),
		Members: memberDTOs,
	}, nil
}

func toGroupResponse(g *entity.PasswordGroup, role entity.GroupRole, memberCount int) dto.GroupResponse {
	return dto.GroupResponse{
		ID:                g.ID,
		Name:              g.Name,
		Description:       g.Description,
		OwnerID:           g.OwnerID,
		AllowMemberCreate: g.AllowMemberCreate,
		AllowMemberEdit:   g.AllowMemberEdit,
		AllowMemberDelete: g.AllowMemberDelete,
		MyRole:            string(role),
		MemberCount:       memberCount,
		CreatedAt:         g.CreatedAt,
		UpdatedAt:         g.UpdatedAt,
	}
}

func toMemberResponse(m *entity.GroupMember) dto.GroupMemberResponse {
	return dto.GroupMemberResponse{
		ID:        m.ID,
		GroupID:   m.GroupID,
		UserID:    m.UserID,
		Email:     m.Email,
		Role:      string(m.Role),
		Status:    string(m.Status),
		CreatedAt: m.CreatedAt,
	}
}
