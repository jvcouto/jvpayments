package workers

import (
	"jvpayments/internal/queue"
	"jvpayments/internal/services"
)

type PaymentWorkerBehavior interface {
	ProcessNextJob()
	RequeueJob(job *queue.PaymentJob) error
	HandleFailedJob(job *queue.PaymentJob, err error)
}

type PaymentWorker struct {
	queue          *queue.RedisPaymentQueue
	paymentService *services.PaymentService
	behavior       PaymentWorkerBehavior
}

// func NewPaymentWorker(queueName string, behavior PaymentWorkerBehavior) *PaymentWorker {
// 	return &PaymentWorker{
// 		queue:          queue.NewRedisPaymentQueue(),
// 		paymentService: services.NewPaymentService(),
// 		behavior:       behavior,
// 	}
// }

// func (pw *PaymentWorker) Start() {
// 	for {
// 		pw.behavior.ProcessNextJob()
// 	}
// }
