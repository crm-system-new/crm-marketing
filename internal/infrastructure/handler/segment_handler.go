package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-marketing/internal/service"
	"github.com/crm-system-new/crm-shared/pkg/httputil"
)

type SegmentHandler struct {
	segmentService *service.SegmentService
}

func NewSegmentHandler(segmentService *service.SegmentService) *SegmentHandler {
	return &SegmentHandler{segmentService: segmentService}
}

func (h *SegmentHandler) CreateSegment(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.segmentService.CreateSegment(r.Context(), req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondCreated(w, resp)
}

func (h *SegmentHandler) GetSegment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.segmentService.GetSegment(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *SegmentHandler) UpdateSegment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req service.UpdateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.segmentService.UpdateSegment(r.Context(), id, req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *SegmentHandler) ListSegments(w http.ResponseWriter, r *http.Request) {
	p := httputil.ParsePagination(r)
	filter := domain.SegmentFilter{Limit: p.Limit, Offset: p.Offset}

	segments, total, err := h.segmentService.ListSegments(r.Context(), filter)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, httputil.ListResponse[*service.SegmentResponse]{
		Items:  segments,
		Total:  total,
		Limit:  p.Limit,
		Offset: p.Offset,
	})
}
