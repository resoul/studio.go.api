package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/resoul/studio.go.api/internal/domain"
	"gorm.io/gorm"
)

// HTTPError holds the HTTP status and the error code string sent to the client.
type HTTPError struct {
	Status  int
	Code    string
	Message string
}

// MapError converts a domain or infrastructure error into an HTTPError.
// Handlers call RespondMapped(c, err) instead of hand-coding status codes.
func MapError(err error) HTTPError {
	switch {
	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		return HTTPError{http.StatusNotFound, "NOT_FOUND", err.Error()}

	case errors.Is(err, domain.ErrConflict):
		return HTTPError{http.StatusConflict, "CONFLICT", err.Error()}

	case errors.Is(err, domain.ErrUnauthorized):
		return HTTPError{http.StatusUnauthorized, "UNAUTHORIZED", err.Error()}

	case errors.Is(err, domain.ErrForbidden):
		return HTTPError{http.StatusForbidden, "FORBIDDEN", err.Error()}

	case errors.Is(err, domain.ErrInvalidInput):
		return HTTPError{http.StatusBadRequest, "INVALID_INPUT", err.Error()}

	case errors.Is(err, domain.ErrInviteExpired):
		return HTTPError{http.StatusGone, "INVITE_EXPIRED", err.Error()}

	case errors.Is(err, domain.ErrOwnerCannotBeRemoved):
		return HTTPError{http.StatusUnprocessableEntity, "OWNER_CANNOT_BE_REMOVED", err.Error()}

	default:
		return HTTPError{http.StatusInternalServerError, "INTERNAL_ERROR", err.Error()}
	}
}

// RespondMapped writes the mapped HTTP error response.
// Use this in handlers instead of manually selecting a status code.
//
//	if err != nil {
//	    utils.RespondMapped(c, err)
//	    return
//	}
func RespondMapped(c *gin.Context, err error) {
	e := MapError(err)
	RespondError(c, e.Status, e.Code, e.Message)
}
