package v1

import (
	"net/http"

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
	subscriptionService *subscriptionuc.Service
	validator           *validator.Validator
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(subscriptionService *subscriptionuc.Service) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService, validator: validator.New()}
}

// ListPlans godoc
// @Summary      List subscription plans
// @Tags         subscription
// @Produce      json
// @Success      200 {object} httputil.Response{data=[]dto.PlanResponse}
// @Router       /plans [get]
func (h *SubscriptionHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, appErr := h.subscriptionService.ListPlans(r.Context())
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.PlansToResponse(plans))
}

// GetCurrent godoc
// @Summary      Get current subscription
// @Tags         subscription
// @Produce      json
// @Success      200 {object} httputil.Response{data=dto.SubscriptionResponse}
// @Security     BearerAuth
// @Router       /subscriptions/me [get]
func (h *SubscriptionHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}
	sub, appErr := h.subscriptionService.GetCurrent(r.Context(), userID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.SubscriptionToResponse(sub))
}

// Subscribe godoc
// @Summary      Create a subscription
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        body body dto.SubscribeRequest true "Subscription data"
// @Success      201 {object} httputil.Response{data=dto.SubscriptionResponse}
// @Security     BearerAuth
// @Router       /subscriptions [post]
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}
	var req dto.SubscribeRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil), errors)
		return
	}

	planID, _ := uuid.Parse(req.PlanID)
	sub, appErr := h.subscriptionService.Subscribe(r.Context(), subscriptionuc.SubscribeInput{UserID: userID, PlanID: planID})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusCreated, mapper.SubscriptionToResponse(sub))
}

// ChangePlan godoc
// @Summary      Change subscription plan
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        body body dto.ChangePlanRequest true "New plan"
// @Success      200 {object} httputil.Response{data=dto.SubscriptionResponse}
// @Security     BearerAuth
// @Router       /subscriptions/me/plan [patch]
func (h *SubscriptionHandler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}
	var req dto.ChangePlanRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}

	planID, _ := uuid.Parse(req.PlanID)
	sub, appErr := h.subscriptionService.ChangePlan(r.Context(), subscriptionuc.ChangePlanInput{UserID: userID, PlanID: planID})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.SubscriptionToResponse(sub))
}

// Cancel godoc
// @Summary      Cancel subscription
// @Tags         subscription
// @Success      204
// @Security     BearerAuth
// @Router       /subscriptions/me/cancel [post]
func (h *SubscriptionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}
	if appErr := h.subscriptionService.Cancel(r.Context(), userID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
