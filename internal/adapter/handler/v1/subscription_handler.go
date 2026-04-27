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

// Assign godoc
// @Summary      Assign subscription
// @Description  Create a new subscription for a user against a plan (admin only)
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        body body dto.AssignSubscriptionRequest true "Subscription assignment"
// @Success      201 {object} httputil.Response{data=dto.SubscriptionResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/subscriptions [post]
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

// Cancel godoc
// @Summary      Cancel subscription
// @Description  Cancel an existing subscription by its ID (admin only)
// @Tags         subscriptions
// @Param        id path string true "Subscription ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/subscriptions/{id} [delete]
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

// GetMySubscription godoc
// @Summary      Get my subscription
// @Description  Return the authenticated user's currently active subscription, or null if none.
// @Tags         subscriptions
// @Produce      json
// @Success      200 {object} httputil.Response{data=dto.SubscriptionResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/subscription [get]
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

// ListByUser godoc
// @Summary      List subscriptions for user
// @Description  Return all subscriptions for the given user (admin only)
// @Tags         subscriptions
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Success      200 {object} httputil.Response{data=[]dto.SubscriptionResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id}/subscriptions [get]
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
