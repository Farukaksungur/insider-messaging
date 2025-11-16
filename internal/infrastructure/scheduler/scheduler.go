package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"insider-messaging/internal/application"
	"insider-messaging/internal/config"
)

var _ application.SchedulerController = (*Scheduler)(nil)

// Scheduler belirli aralıklarla mesaj gönderme işlemini çalıştırır
type Scheduler struct {
	uc      *application.SendBatchUseCase
	cfg     *config.Config
	ticker  *time.Ticker
	stopCh  chan struct{}
	running bool
	mu      sync.Mutex
	wg      sync.WaitGroup
}

// NewScheduler yeni bir scheduler oluşturur
func NewScheduler(uc *application.SendBatchUseCase, cfg *config.Config) *Scheduler {
	return &Scheduler{uc: uc, cfg: cfg}
}

// Start scheduler'ı başlatır, zaten çalışıyorsa bir şey yapmaz
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}

	interval := time.Duration(s.cfg.ScheduleSec) * time.Second
	if interval <= 0 {
		interval = 120 * time.Second
	}

	s.ticker = time.NewTicker(interval)
	s.stopCh = make(chan struct{})
	s.running = true
	s.wg.Add(1)
	go s.loop()
}

// loop scheduler'ın ana döngüsü, ticker'a göre mesaj gönderme işlemini çalıştırır
func (s *Scheduler) loop() {
	defer s.wg.Done()
	for {
		select {
		case <-s.ticker.C:
			timeout := time.Duration(s.cfg.WebhookTimeoutSeconds+10) * time.Second
			if timeout <= 0 {
				timeout = 30 * time.Second
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			if err := s.uc.Execute(ctx); err != nil {
				log.Printf("sendbatch err: %v", err)
			}
			cancel()
		case <-s.stopCh:
			return
		}
	}
}

// Stop scheduler'ı durdurur ve tüm işlemlerin bitmesini bekler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.ticker.Stop()
	close(s.stopCh)
	s.wg.Wait()
	s.running = false
}

// IsRunning scheduler'ın çalışıp çalışmadığını döndürür
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
