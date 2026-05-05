package v1

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	billinguc "github.com/zenfulcode/zencial/internal/usecase/billing"
)

type BillingHandler struct {
	billingService *billinguc.Service
	validator      *validator.Validator
}

func NewBillingHandler(billingService *billinguc.Service) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
		validator:      validator.New(),
	}
}

func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	var req dto.CreateCheckoutSessionRequest
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

	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeValidationFailed, "invalid plan ID")
		return
	}

	session, appErr := h.billingService.CreateCheckoutSession(r.Context(), billinguc.CheckoutInput{
		UserID:     userID,
		PlanID:     planID,
		SuccessURL: req.SuccessURL,
		CancelURL:  req.CancelURL,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BillingSessionToResponse(session))
}

func (h *BillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	var req dto.CreatePortalSessionRequest
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

	session, appErr := h.billingService.CreatePortalSession(r.Context(), billinguc.PortalInput{
		UserID:    userID,
		ReturnURL: req.ReturnURL,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BillingSessionToResponse(session))
}

func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	invoices, appErr := h.billingService.ListInvoices(r.Context(), userID, 12)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BillingInvoicesToResponse(invoices))
}

func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid webhook body")
		return
	}

	appErr := h.billingService.HandleWebhook(r.Context(), payload, r.Header.Get("Stripe-Signature"))
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
