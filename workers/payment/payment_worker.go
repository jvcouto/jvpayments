package workers

import (
	"jvpayments/queue"
	"jvpayments/services"
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

func NewPaymentWorker(queueName string, behavior PaymentWorkerBehavior) *PaymentWorker {
	return &PaymentWorker{
		queue:          queue.NewRedisPaymentQueue(queueName),
		paymentService: services.NewPaymentService(),
		behavior:       behavior,
	}
}

func (pw *PaymentWorker) Start() {
	for {
		pw.behavior.ProcessNextJob()
	}
}
