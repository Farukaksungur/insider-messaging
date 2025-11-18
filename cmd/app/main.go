package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"insider-messaging/internal/application"
	"insider-messaging/internal/config"
	"insider-messaging/internal/infrastructure/cache"
	db "insider-messaging/internal/infrastructure/db"
	"insider-messaging/internal/infrastructure/scheduler"
	"insider-messaging/internal/infrastructure/sender"
	"insider-messaging/internal/presentation/api"
)

// @title           Insider Messaging API
// @version         1.0
// @description     Automatic message sending system with webhook integration
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @schemes   http https

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	gormDB, err := db.NewMySQL(cfg)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}

	redisClient := cache.NewRedis(cfg)

	msgRepo := db.NewMySQLMessageRepository(gormDB)
	webSender := sender.NewWebhookSender(cfg)
	sendBatchUC := application.NewSendBatchUseCase(msgRepo, webSender, redisClient, cfg)
	sched := scheduler.NewScheduler(sendBatchUC, cfg)

	router := api.NewRouter(sched, msgRepo, cfg)
	srv := api.NewServer(cfg, router)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("http server stopped: %v", err)
		}
	}()
	log.Printf("server started on :%s", cfg.Port)
	<-stop
	log.Println("shutdown signal received")
	sched.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("exited cleanly")
}
