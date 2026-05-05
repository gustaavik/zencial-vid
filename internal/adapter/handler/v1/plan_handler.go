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

// Create godoc
// @Summary      Create plan
// @Description  Create a new subscription plan (admin only)
// @Tags         plans
// @Accept       json
// @Produce      json
// @Param        body body dto.CreatePlanRequest true "Plan data"
// @Success      201 {object} httputil.Response{data=dto.PlanResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /plans [post]
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
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Level:         req.Level,
		StripePriceID: req.StripePriceID,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.PlanToResponse(plan))
}

// GetByID godoc
// @Summary      Get plan by ID
// @Description  Return a single plan by its UUID (admin only)
// @Tags         plans
// @Produce      json
// @Param        id path string true "Plan ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.PlanResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /plans/{id} [get]
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

// List godoc
// @Summary      List all plans (admin)
// @Description  Return a paginated list of all plans, including inactive ones (admin only)
// @Tags         plans
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        sort query string false "Sort field (e.g. created_at,-level)"
// @Success      200 {object} httputil.Response{data=[]dto.PlanResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/plans [get]
func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.PlanFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	plans, total, appErr := h.planService.List(r.Context(), &fs)
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

// ListActive godoc
// @Summary      List active plans
// @Description  Return all currently active subscription plans
// @Tags         plans
// @Produce      json
// @Success      200 {object} httputil.Response{data=[]dto.PlanResponse}
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /plans [get]
func (h *PlanHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	plans, appErr := h.planService.ListActive(r.Context())
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.PlansToResponse(plans))
}

// Update godoc
// @Summary      Update plan
// @Description  Update an existing subscription plan (admin only)
// @Tags         plans
// @Accept       json
// @Produce      json
// @Param        id path string true "Plan ID" format(uuid)
// @Param        body body dto.UpdatePlanRequest true "Plan fields to update"
// @Success      200 {object} httputil.Response{data=dto.PlanResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /plans/{id} [put]
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
		ID:            id,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Level:         req.Level,
		StripePriceID: req.StripePriceID,
		IsActive:      req.IsActive,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.PlanToResponse(plan))
}

// Delete godoc
// @Summary      Delete plan
// @Description  Soft-delete a subscription plan (admin only)
// @Tags         plans
// @Param        id path string true "Plan ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /plans/{id} [delete]
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
