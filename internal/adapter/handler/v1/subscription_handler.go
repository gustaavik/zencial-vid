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

// AdminListSubscriptions godoc
// @Summary      List all subscriptions (admin)
// @Tags         admin
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.AdminSubscriptionResponse}
// @Security     BearerAuth
// @Router       /admin/subscriptions [get]
func (h *SubscriptionHandler) AdminListSubscriptions(w http.ResponseWriter, r *http.Request) {
	page := httputil.QueryInt(r, "page", 1)
	perPage := httputil.QueryInt(r, "per_page", 20)

	subs, total, appErr := h.subscriptionService.AdminListSubscriptions(r.Context(), page, perPage)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	meta := &httputil.Meta{
		Page: page, PerPage: perPage, Total: total,
		TotalPages: int(total) / perPage,
	}
	if int(total)%perPage > 0 {
		meta.TotalPages++
	}
	httputil.SuccessWithMeta(w, mapper.AdminSubscriptionsToResponse(subs), meta)
}

// AdminGetUserSubscription godoc
// @Summary      Get a user's subscription (admin)
// @Tags         admin
// @Produce      json
// @Param        userId path string true "User ID"
// @Success      200 {object} httputil.Response{data=dto.SubscriptionResponse}
// @Security     BearerAuth
// @Router       /admin/subscriptions/user/{userId} [get]
func (h *SubscriptionHandler) AdminGetUserSubscription(w http.ResponseWriter, r *http.Request) {
	userID, err := httputil.URLParamUUID(r, "userId")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid user ID")
		return
	}

	sub, appErr := h.subscriptionService.AdminGetUserSubscription(r.Context(), userID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.SubscriptionToResponse(sub))
}

// AdminChangePlan godoc
// @Summary      Change a subscription's plan (admin)
// @Tags         admin
// @Accept       json
// @Param        id path string true "Subscription ID"
// @Param        body body dto.AdminChangePlanRequest true "New plan"
// @Success      200 {object} httputil.Response{data=dto.SubscriptionResponse}
// @Security     BearerAuth
// @Router       /admin/subscriptions/{id}/plan [patch]
func (h *SubscriptionHandler) AdminChangePlan(w http.ResponseWriter, r *http.Request) {
	subID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid subscription ID")
		return
	}

	var req dto.AdminChangePlanRequest
	if decodeErr := httputil.DecodeJSON(r, &req); decodeErr != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil), errors)
		return
	}

	planID, _ := uuid.Parse(req.PlanID)
	sub, appErr := h.subscriptionService.AdminChangePlan(r.Context(), subscriptionuc.AdminChangePlanInput{
		SubscriptionID: subID,
		PlanID:         planID,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.SubscriptionToResponse(sub))
}

// AdminCreateSubscription godoc
// @Summary      Create a subscription for a user (admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        body body dto.AdminCreateSubscriptionRequest true "Subscription data"
// @Success      201 {object} httputil.Response{data=dto.SubscriptionResponse}
// @Security     BearerAuth
// @Router       /admin/subscriptions [post]
func (h *SubscriptionHandler) AdminCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req dto.AdminCreateSubscriptionRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil), errors)
		return
	}

	userID, _ := uuid.Parse(req.UserID)
	planID, _ := uuid.Parse(req.PlanID)
	sub, appErr := h.subscriptionService.AdminCreateSubscription(r.Context(), subscriptionuc.AdminCreateSubscriptionInput{
		UserID: userID,
		PlanID: planID,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusCreated, mapper.SubscriptionToResponse(sub))
}

// AdminReactivateSubscription godoc
// @Summary      Reactivate a canceled subscription (admin)
// @Tags         admin
// @Param        id path string true "Subscription ID"
// @Success      200 {object} httputil.Response{data=dto.AdminSubscriptionResponse}
// @Security     BearerAuth
// @Router       /admin/subscriptions/{id}/reactivate [post]
func (h *SubscriptionHandler) AdminReactivateSubscription(w http.ResponseWriter, r *http.Request) {
	subID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid subscription ID")
		return
	}

	sub, appErr := h.subscriptionService.AdminReactivateSubscription(r.Context(), subID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.AdminSubscriptionToResponse(sub))
}

// AdminCancelSubscription godoc
// @Summary      Cancel a subscription (admin)
// @Tags         admin
// @Param        id path string true "Subscription ID"
// @Success      204
// @Security     BearerAuth
// @Router       /admin/subscriptions/{id}/cancel [post]
func (h *SubscriptionHandler) AdminCancelSubscription(w http.ResponseWriter, r *http.Request) {
	subID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid subscription ID")
		return
	}

	if appErr := h.subscriptionService.AdminCancelSubscription(r.Context(), subID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
