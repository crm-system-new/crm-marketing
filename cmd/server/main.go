package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crm-system-new/crm-marketing/internal/infrastructure/handler"
	inframsg "github.com/crm-system-new/crm-marketing/internal/infrastructure/messaging"
	"github.com/crm-system-new/crm-marketing/internal/infrastructure/postgres"
	"github.com/crm-system-new/crm-marketing/internal/service"
	"github.com/crm-system-new/crm-shared/pkg/audit"
	"github.com/crm-system-new/crm-shared/pkg/auth"
	"github.com/crm-system-new/crm-shared/pkg/config"
	"github.com/crm-system-new/crm-shared/pkg/health"
	"github.com/crm-system-new/crm-shared/pkg/idempotency"
	sharedotel "github.com/crm-system-new/crm-shared/pkg/otel"
	"github.com/crm-system-new/crm-shared/pkg/outbox"
	sharedpg "github.com/crm-system-new/crm-shared/pkg/postgres"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load("marketing")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	shutdown, err := sharedotel.InitTracer(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	if err != nil {
		log.Printf("WARN: failed to init tracer: %v", err)
	} else {
		defer shutdown(ctx)
	}

	pool, err := sharedpg.NewPool(ctx, sharedpg.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Database: cfg.DBName,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	publisher, err := inframsg.NewMarketingPublisher(cfg.NatsURL)
	if err != nil {
		log.Fatalf("connect to nats: %v", err)
	}
	defer publisher.Close()

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 15*time.Minute, 7*24*time.Hour)

	// Production infrastructure
	outboxStore := outbox.NewPgStore(pool)
	auditLogger := audit.NewLogger(pool)
	idempotencyStore := idempotency.NewStore(pool)
	healthChecker := health.NewChecker(pool, cfg.NatsURL)

	// Wire dependencies
	campaignRepo := postgres.NewCampaignRepository(pool)
	segmentRepo := postgres.NewSegmentRepository(pool)
	subscriberRepo := postgres.NewSubscriberRepository(pool)

	campaignService := service.NewCampaignService(campaignRepo, pool, outboxStore, auditLogger)
	segmentService := service.NewSegmentService(segmentRepo, pool, outboxStore, auditLogger)
	subscriberService := service.NewSubscriberService(subscriberRepo, pool, outboxStore, auditLogger)

	campaignHandler := handler.NewCampaignHandler(campaignService)
	segmentHandler := handler.NewSegmentHandler(segmentService)
	subscriberHandler := handler.NewSubscriberHandler(subscriberService)

	// Start outbox relay (polls outbox table and publishes to NATS)
	relay := outbox.NewRelay(outboxStore, publisher, 2*time.Second, 100)
	relay.Start()
	defer relay.Stop()

	// Start sales event subscriber (consumes lead.created, customer.created)
	salesSubscriber, err := inframsg.NewSalesEventSubscriber(cfg.NatsURL, subscriberRepo, idempotencyStore)
	if err != nil {
		log.Printf("WARN: failed to start sales event subscriber: %v", err)
	} else {
		defer salesSubscriber.Close()
	}

	router := handler.NewRouter(campaignHandler, segmentHandler, subscriberHandler, jwtManager, healthChecker)

	addr := fmt.Sprintf(":%d", cfg.ServicePort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Marketing service listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down marketing service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
