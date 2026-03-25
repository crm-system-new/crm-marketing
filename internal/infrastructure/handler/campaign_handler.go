package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-marketing/internal/service"
	"github.com/crm-system-new/crm-shared/pkg/httputil"
)

type CampaignHandler struct {
	campaignService *service.CampaignService
}

func NewCampaignHandler(campaignService *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{campaignService: campaignService}
}

func (h *CampaignHandler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req service.CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.campaignService.CreateCampaign(r.Context(), req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondCreated(w, resp)
}

func (h *CampaignHandler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.campaignService.GetCampaign(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *CampaignHandler) LaunchCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.campaignService.LaunchCampaign(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *CampaignHandler) PauseCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.campaignService.PauseCampaign(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *CampaignHandler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	p := httputil.ParsePagination(r)
	filter := domain.CampaignFilter{Limit: p.Limit, Offset: p.Offset}

	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.CampaignStatus(v)
		filter.Status = &status
	}
	if v := r.URL.Query().Get("channel"); v != "" {
		filter.Channel = &v
	}
	if v := r.URL.Query().Get("segment_id"); v != "" {
		filter.SegmentID = &v
	}

	campaigns, total, err := h.campaignService.ListCampaigns(r.Context(), filter)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, httputil.ListResponse[*service.CampaignResponse]{
		Items:  campaigns,
		Total:  total,
		Limit:  p.Limit,
		Offset: p.Offset,
	})
}
