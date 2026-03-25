package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/crm-system-new/crm-shared/pkg/auth"
)

func NewRouter(campaignH *CampaignHandler, segmentH *SegmentHandler, subscriberH *SubscriberHandler, jwtManager *auth.JWTManager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware(jwtManager))

		// Campaigns
		r.Route("/campaigns", func(r chi.Router) {
			r.Post("/", campaignH.CreateCampaign)
			r.Get("/", campaignH.ListCampaigns)
			r.Get("/{id}", campaignH.GetCampaign)
			r.Post("/{id}/launch", campaignH.LaunchCampaign)
			r.Post("/{id}/pause", campaignH.PauseCampaign)
		})

		// Segments
		r.Route("/segments", func(r chi.Router) {
			r.Post("/", segmentH.CreateSegment)
			r.Get("/", segmentH.ListSegments)
			r.Get("/{id}", segmentH.GetSegment)
			r.Put("/{id}", segmentH.UpdateSegment)
		})

		// Subscribers
		r.Route("/subscribers", func(r chi.Router) {
			r.Post("/", subscriberH.Subscribe)
			r.Get("/", subscriberH.ListSubscribers)
			r.Get("/{id}", subscriberH.GetSubscriber)
			r.Post("/{id}/unsubscribe", subscriberH.Unsubscribe)
			r.Put("/{id}/preferences", subscriberH.UpdatePreferences)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return r
}
