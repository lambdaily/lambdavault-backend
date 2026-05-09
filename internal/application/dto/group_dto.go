package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateGroupRequest is the payload to create a password group. The creator
// becomes the admin/owner automatically.
type CreateGroupRequest struct {
	Name              string             `json:"name" validate:"required,min=1,max=100"`
	Description       string             `json:"description" validate:"omitempty,max=500"`
	AllowMemberCreate bool               `json:"allow_member_create"`
	AllowMemberEdit   bool               `json:"allow_member_edit"`
	AllowMemberDelete bool               `json:"allow_member_delete"`
	Members           []GroupInviteEntry `json:"members" validate:"omitempty,dive"`
}

// UpdateGroupRequest updates the editable settings of a group.
type UpdateGroupRequest struct {
	Name              string `json:"name" validate:"required,min=1,max=100"`
	Description       string `json:"description" validate:"omitempty,max=500"`
	AllowMemberCreate bool   `json:"allow_member_create"`
	AllowMemberEdit   bool   `json:"allow_member_edit"`
	AllowMemberDelete bool   `json:"allow_member_delete"`
}

// GroupInviteEntry pairs an email with the role to assign to that member.
type GroupInviteEntry struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin editor viewer"`
}

// AddMembersRequest invites a list of emails to a group with explicit roles.
type AddMembersRequest struct {
	Members []GroupInviteEntry `json:"members" validate:"required,min=1,dive"`
}

// UpdateMemberRoleRequest changes the role of an existing member.
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin editor viewer"`
}

// CreateGroupPasswordRequest creates a password inside a group.
type CreateGroupPasswordRequest struct {
	SiteName string `json:"site_name" validate:"required,min=1,max=255"`
	SiteURL  string `json:"site_url" validate:"omitempty,max=500"`
	Username string `json:"username" validate:"required,min=1,max=255"`
	Password string `json:"password" validate:"required,min=1,max=500"`
	Notes    string `json:"notes" validate:"omitempty,max=1000"`
	Category string `json:"category" validate:"omitempty,max=100"`
}

// AddExistingGroupPasswordRequest shares an already-created personal password
// with a group by reference so it remains in the personal vault too.
type AddExistingGroupPasswordRequest struct {
	PasswordID uuid.UUID `json:"password_id" validate:"required"`
}

// GroupResponse is the public representation of a group.
type GroupResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	OwnerID           uuid.UUID `json:"owner_id"`
	AllowMemberCreate bool      `json:"allow_member_create"`
	AllowMemberEdit   bool      `json:"allow_member_edit"`
	AllowMemberDelete bool      `json:"allow_member_delete"`
	MyRole            string    `json:"my_role"`
	MemberCount       int       `json:"member_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// GroupMemberResponse is the public representation of a group member.
type GroupMemberResponse struct {
	ID        uuid.UUID  `json:"id"`
	GroupID   uuid.UUID  `json:"group_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

// GroupListResponse is a list of groups.
type GroupListResponse struct {
	Groups []GroupResponse `json:"groups"`
	Total  int             `json:"total"`
}

// GroupMembersResponse is a list of members for a group.
type GroupMembersResponse struct {
	Members []GroupMemberResponse `json:"members"`
	Total   int                   `json:"total"`
}

// GroupDetailResponse bundles the group, its members and the caller's role.
type GroupDetailResponse struct {
	Group   GroupResponse         `json:"group"`
	Members []GroupMemberResponse `json:"members"`
}
