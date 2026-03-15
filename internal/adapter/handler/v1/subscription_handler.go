package v1

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	subscriptionuc "github.com/zenfulcode/zencial/internal/usecase/subscription"
)

// SubscriptionHandler handles subscription HTTP requests.
type SubscriptionHandler struct {
	subService *subscriptionuc.Service
	validator  *validator.Validator
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(subService *subscriptionuc.Service) *SubscriptionHandler {
	return &SubscriptionHandler{
		subService: subService,
		validator:  validator.New(),
	}
}

// Assign creates a new subscription for a user (admin).
func (h *SubscriptionHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req dto.AssignSubscriptionRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil),
			errors,
		)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeValidationFailed, "invalid user ID")
		return
	}

	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeValidationFailed, "invalid plan ID")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			httputil.BadRequest(w, apperror.CodeValidationFailed, "invalid expires_at format, expected RFC3339")
			return
		}
		expiresAt = &t
	}

	sub, appErr := h.subService.Assign(r.Context(), subscriptionuc.AssignInput{
		UserID:    userID,
		PlanID:    planID,
		ExpiresAt: expiresAt,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.SubscriptionToResponse(sub))
}

// Cancel cancels a subscription (admin).
func (h *SubscriptionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid subscription ID")
		return
	}

	appErr := h.subService.Cancel(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMySubscription returns the authenticated user's active subscription.
func (h *SubscriptionHandler) GetMySubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	swp, appErr := h.subService.GetActiveByUserID(r.Context(), userID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	if swp == nil {
		httputil.Success(w, http.StatusOK, nil)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.SubscriptionWithPlanToResponse(swp))
}

// ListByUser returns all subscriptions for a user (admin).
func (h *SubscriptionHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid user ID")
		return
	}

	subs, appErr := h.subService.ListByUserID(r.Context(), userID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.SubscriptionsToResponse(subs))
}
