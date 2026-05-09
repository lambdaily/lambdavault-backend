package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Notifier emits external notifications about user-visible vault events.
//
// Implementations must treat failures as non-fatal to the core password flow;
// the API should still succeed even if the notification provider is down.
type Notifier interface {
	NotifyPasswordCreated(ctx context.Context, event PasswordCreatedEvent) error
	NotifyPasswordShared(ctx context.Context, event PasswordSharedEvent) error
}

type PasswordCreatedEvent struct {
	PasswordID      uuid.UUID `json:"password_id"`
	OwnerID         uuid.UUID `json:"owner_id"`
	OwnerEmail      string    `json:"owner_email"`
	OwnerPhone      string    `json:"owner_phone,omitempty"`
	SiteName        string    `json:"site_name"`
	SiteURL         string    `json:"site_url,omitempty"`
	Username        string    `json:"username"`
	Category        string    `json:"category,omitempty"`
	OccurredAt      time.Time `json:"occurred_at"`
	RecipientEmails []string  `json:"recipient_emails"`
	RecipientPhones []string  `json:"recipient_phones"`
}

type PasswordSharedEvent struct {
	PasswordID      uuid.UUID `json:"password_id"`
	GroupID         uuid.UUID `json:"group_id"`
	GroupName       string    `json:"group_name"`
	SharedByUserID  uuid.UUID `json:"shared_by_user_id"`
	SharedByEmail   string    `json:"shared_by_email"`
	SharedByPhone   string    `json:"shared_by_phone,omitempty"`
	SiteName        string    `json:"site_name"`
	SiteURL         string    `json:"site_url,omitempty"`
	Username        string    `json:"username"`
	Category        string    `json:"category,omitempty"`
	OccurredAt      time.Time `json:"occurred_at"`
	RecipientEmails []string  `json:"recipient_emails"`
	RecipientPhones []string  `json:"recipient_phones"`
}

type NoopNotifier struct{}

func (NoopNotifier) NotifyPasswordCreated(context.Context, PasswordCreatedEvent) error {
	return nil
}

func (NoopNotifier) NotifyPasswordShared(context.Context, PasswordSharedEvent) error {
	return nil
}
