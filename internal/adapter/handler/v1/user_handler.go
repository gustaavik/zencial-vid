package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
)

// UserHandler handles user profile HTTP requests.
type UserHandler struct {
	validator *validator.Validator
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler() *UserHandler {
	return &UserHandler{
		validator: validator.New(),
	}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	httputil.NotFound(w, "NOT_FOUND", "not implemented")
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	httputil.NotFound(w, "NOT_FOUND", "not implemented")
}

func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	httputil.NotFound(w, "NOT_FOUND", "not implemented")
}
