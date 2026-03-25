package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-marketing/internal/service"
	"github.com/crm-system-new/crm-shared/pkg/httputil"
)

type SubscriberHandler struct {
	subscriberService *service.SubscriberService
}

func NewSubscriberHandler(subscriberService *service.SubscriberService) *SubscriberHandler {
	return &SubscriberHandler{subscriberService: subscriberService}
}

func (h *SubscriberHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req service.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.subscriberService.Subscribe(r.Context(), req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondCreated(w, resp)
}

func (h *SubscriberHandler) GetSubscriber(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.subscriberService.GetSubscriber(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *SubscriberHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.subscriberService.Unsubscribe(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *SubscriberHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req service.UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.subscriberService.UpdatePreferences(r.Context(), id, req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *SubscriberHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	p := httputil.ParsePagination(r)
	filter := domain.SubscriberFilter{Limit: p.Limit, Offset: p.Offset}

	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.SubscriberStatus(v)
		filter.Status = &status
	}
	if v := r.URL.Query().Get("email"); v != "" {
		filter.Email = &v
	}

	subscribers, total, err := h.subscriberService.ListSubscribers(r.Context(), filter)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, httputil.ListResponse[*service.SubscriberResponse]{
		Items:  subscribers,
		Total:  total,
		Limit:  p.Limit,
		Offset: p.Offset,
	})
}
