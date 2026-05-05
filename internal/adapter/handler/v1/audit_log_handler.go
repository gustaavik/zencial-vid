package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	audituc "github.com/zenfulcode/zencial/internal/usecase/audit"
)

// AuditLogHandler handles admin audit log HTTP requests.
type AuditLogHandler struct {
	service *audituc.Service
}

// NewAuditLogHandler creates a new AuditLogHandler.
func NewAuditLogHandler(service *audituc.Service) *AuditLogHandler {
	return &AuditLogHandler{service: service}
}

// List godoc
// @Summary      List audit logs
// @Description  Return a paginated list of admin audit log entries (admin only)
// @Tags         audit-logs
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(25)
// @Param        sort query string false "Sort field (default: -occurred_at)"
// @Success      200 {object} httputil.Response{data=[]dto.AuditLogResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/audit-logs [get]
func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.AuditLogFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	logs, total, appErr := h.service.List(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.AuditLogsToResponse(logs), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}
