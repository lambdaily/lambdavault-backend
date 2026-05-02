package errors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrPasswordNotFound = errors.New("password not found")
	ErrAccessDenied     = errors.New("access denied")

	ErrGroupNotFound         = errors.New("group not found")
	ErrGroupMemberNotFound   = errors.New("group member not found")
	ErrGroupMemberExists     = errors.New("group member already exists")
	ErrGroupOwnerImmutable   = errors.New("the group owner cannot be removed or demoted")
	ErrInvalidGroupRole      = errors.New("invalid group role")
	ErrCannotInviteSelf      = errors.New("cannot invite yourself to a group")
	ErrInsufficientGroupRole = errors.New("you do not have permission to perform this action in the group")

	ErrEncryptionFailed = errors.New("encryption failed")
	ErrDecryptionFailed = errors.New("decryption failed")

	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenMalformed    = errors.New("token malformed")
	ErrMissingAuthHeader = errors.New("missing authorization header")

	ErrInternalServer = errors.New("internal server error")
)

type DomainError struct {
	Err     error
	Message string
	Code    string
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewDomainError(err error, message, code string) *DomainError {
	return &DomainError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}
