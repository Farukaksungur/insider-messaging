package application

// SchedulerController scheduler kontrolü için interface
type SchedulerController interface {
	Start()
	Stop()
	IsRunning() bool
}
