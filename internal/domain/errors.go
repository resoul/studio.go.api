package domain

import "errors"

// Sentinel errors used across all layers.
// The transport layer maps these to HTTP status codes via MapError.
var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrConflict is returned when an operation violates a uniqueness constraint.
	ErrConflict = errors.New("conflict")

	// ErrUnauthorized is returned when the caller lacks the required identity.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the caller is authenticated but not permitted.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidInput is returned when caller-supplied data fails validation.
	ErrInvalidInput = errors.New("invalid input")

	// ErrInviteExpired is returned when an invite token is past its expiry date.
	ErrInviteExpired = errors.New("invite expired")

	// ErrOwnerCannotBeRemoved is returned when attempting to remove a workspace owner.
	ErrOwnerCannotBeRemoved = errors.New("cannot remove workspace owner")
)
