package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// GroupRole represents the role of a member inside a password group.
type GroupRole string

const (
	GroupRoleAdmin  GroupRole = "admin"
	GroupRoleEditor GroupRole = "editor"
	GroupRoleViewer GroupRole = "viewer"
)

func (r GroupRole) IsValid() bool {
	switch r {
	case GroupRoleAdmin, GroupRoleEditor, GroupRoleViewer:
		return true
	}
	return false
}

// GroupMemberStatus represents the lifecycle status of a group membership.
type GroupMemberStatus string

const (
	GroupMemberStatusPending GroupMemberStatus = "pending"
	GroupMemberStatusActive  GroupMemberStatus = "active"
)

// PasswordGroup is a shared collection of passwords accessible by a team.
//
// The owner is always implicitly an administrator. The boolean flags below
// describe what non-admin members (editor role) are allowed to do. Viewers
// can only read passwords. Admins can always do everything (including
// managing members and group settings).
type PasswordGroup struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"not null"`
	Description string
	OwnerID     uuid.UUID `gorm:"type:uuid;index;not null"`

	// Permission flags chosen by the admin/creator. They gate the actions that
	// editor members are allowed to perform on passwords inside the group.
	AllowMemberCreate bool `gorm:"not null;default:false"`
	AllowMemberEdit   bool `gorm:"not null;default:false"`
	AllowMemberDelete bool `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Owner   *User          `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
	Members []*GroupMember `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
}

func (g *PasswordGroup) TableName() string {
	return "password_groups"
}

// NewPasswordGroup creates a new password group with the given owner. The
// owner is responsible for membership and permission management.
func NewPasswordGroup(ownerID uuid.UUID, name, description string, allowCreate, allowEdit, allowDelete bool) *PasswordGroup {
	return &PasswordGroup{
		ID:                uuid.New(),
		Name:              name,
		Description:       description,
		OwnerID:           ownerID,
		AllowMemberCreate: allowCreate,
		AllowMemberEdit:   allowEdit,
		AllowMemberDelete: allowDelete,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// UpdateSettings updates the editable fields of the group.
func (g *PasswordGroup) UpdateSettings(name, description string, allowCreate, allowEdit, allowDelete bool) {
	g.Name = name
	g.Description = description
	g.AllowMemberCreate = allowCreate
	g.AllowMemberEdit = allowEdit
	g.AllowMemberDelete = allowDelete
	g.UpdatedAt = time.Now()
}

// CanCreatePasswords reports whether a member with the given role can create
// passwords inside this group.
func (g *PasswordGroup) CanCreatePasswords(role GroupRole) bool {
	switch role {
	case GroupRoleAdmin:
		return true
	case GroupRoleEditor:
		return g.AllowMemberCreate
	}
	return false
}

// CanEditPasswords reports whether a member with the given role can edit
// passwords inside this group.
func (g *PasswordGroup) CanEditPasswords(role GroupRole) bool {
	switch role {
	case GroupRoleAdmin:
		return true
	case GroupRoleEditor:
		return g.AllowMemberEdit
	}
	return false
}

// CanDeletePasswords reports whether a member with the given role can delete
// passwords inside this group.
func (g *PasswordGroup) CanDeletePasswords(role GroupRole) bool {
	switch role {
	case GroupRoleAdmin:
		return true
	case GroupRoleEditor:
		return g.AllowMemberDelete
	}
	return false
}

// GroupMember represents a user (or pending invitation) inside a password
// group. When the user already exists in the system UserID is set; otherwise
// the membership is created as pending and is linked by email when the user
// signs up.
type GroupMember struct {
	ID      uuid.UUID  `gorm:"type:uuid;primaryKey"`
	GroupID uuid.UUID  `gorm:"type:uuid;index;not null"`
	UserID  *uuid.UUID `gorm:"type:uuid;index"`
	Email   string     `gorm:"index;not null"`
	Role    GroupRole  `gorm:"type:varchar(16);not null"`
	Status  GroupMemberStatus `gorm:"type:varchar(16);not null"`

	InvitedByID uuid.UUID `gorm:"type:uuid;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Group *PasswordGroup `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
	User  *User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (m *GroupMember) TableName() string {
	return "group_members"
}

// NewGroupMember creates a new membership. If user is non-nil the member is
// considered active immediately; otherwise the invitation stays pending until
// the user signs up with the matching email.
func NewGroupMember(groupID, invitedByID uuid.UUID, email string, role GroupRole, user *User) *GroupMember {
	email = strings.ToLower(strings.TrimSpace(email))
	m := &GroupMember{
		ID:          uuid.New(),
		GroupID:     groupID,
		Email:       email,
		Role:        role,
		Status:      GroupMemberStatusPending,
		InvitedByID: invitedByID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if user != nil {
		m.UserID = &user.ID
		m.Status = GroupMemberStatusActive
	}
	return m
}

// Activate links a pending invitation to a real user.
func (m *GroupMember) Activate(userID uuid.UUID) {
	m.UserID = &userID
	m.Status = GroupMemberStatusActive
	m.UpdatedAt = time.Now()
}

// ChangeRole updates the role of the member.
func (m *GroupMember) ChangeRole(role GroupRole) {
	m.Role = role
	m.UpdatedAt = time.Now()
}
