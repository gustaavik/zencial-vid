package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	planuc "github.com/zenfulcode/zencial/internal/usecase/plan"
)

// PlanHandler handles plan HTTP requests.
type PlanHandler struct {
	planService *planuc.Service
	validator   *validator.Validator
}

// NewPlanHandler creates a new PlanHandler.
func NewPlanHandler(planService *planuc.Service) *PlanHandler {
	return &PlanHandler{
		planService: planService,
		validator:   validator.New(),
	}
}

// Create creates a new plan.
func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePlanRequest
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

	plan, appErr := h.planService.Create(r.Context(), planuc.CreateInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Level:       req.Level,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.PlanToResponse(plan))
}

// GetByID returns a plan by ID.
func (h *PlanHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid plan ID")
		return
	}

	plan, appErr := h.planService.GetByID(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.PlanToResponse(plan))
}

// List returns a paginated list of all plans (admin).
func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.PlanFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	plans, total, appErr := h.planService.List(r.Context(), fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.PlansToResponse(plans), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// ListActive returns all active plans (public).
func (h *PlanHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	plans, appErr := h.planService.ListActive(r.Context())
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.PlansToResponse(plans))
}

// Update updates an existing plan.
func (h *PlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid plan ID")
		return
	}

	var req dto.UpdatePlanRequest
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

	plan, appErr := h.planService.Update(r.Context(), planuc.UpdateInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Level:       req.Level,
		IsActive:    req.IsActive,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.PlanToResponse(plan))
}

// Delete soft-deletes a plan.
func (h *PlanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid plan ID")
		return
	}

	appErr := h.planService.Delete(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
