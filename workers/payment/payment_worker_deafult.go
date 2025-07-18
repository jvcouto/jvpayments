package workers

import (
	"jvpayments/config"
	"jvpayments/queue"
	"jvpayments/services"
	"log"
	"time"
)

type DefaultWorkerBehavior struct {
	queue          *queue.RedisPaymentQueue
	paymentService *services.PaymentService
}

func NewDefaultWorkerBehavior(queue *queue.RedisPaymentQueue, paymentService *services.PaymentService) *DefaultWorkerBehavior {
	return &DefaultWorkerBehavior{
		queue:          queue,
		paymentService: paymentService,
	}
}

func (pw *DefaultWorkerBehavior) ProcessNextJob() {
	config := config.LoadConfig()
	job, err := pw.queue.ConsumePaymentJob()
	if err != nil {
		log.Printf("Error consuming payment job: %v", err)
		return
	}

	log.Printf("Processing payment job: %s", job.ID)

	err = pw.paymentService.ProcessPayment(job.PaymentData, config.PaymentApiUrl)
	if err != nil {
		log.Printf("Error processing payment job %s: %v", job.ID, err)

		if job.RetryCount < job.MaxRetries {
			job.RetryCount++
			log.Printf("Retrying payment job %s (attempt %d/%d)", job.ID, job.RetryCount, job.MaxRetries)
			pw.RequeueJob(job)
		} else {
			log.Printf("Payment job %s failed after %d retries", job.ID, job.MaxRetries)
			pw.HandleFailedJob(job, err)
		}
		return
	}

	log.Printf("Successfully processed payment job %s: %s", job.ID)
}

func (pw *DefaultWorkerBehavior) RequeueJob(job *queue.PaymentJob) error {
	delay := time.Duration(job.RetryCount) * time.Second
	time.Sleep(delay)

	return pw.queue.RequeueJob(job)
}

func (pw *DefaultWorkerBehavior) HandleFailedJob(job *queue.PaymentJob, err error) {
	// TODO: Implement failed job handling
	// - Log to error queue
	// - Send notification
	// - Store in database for manual review
	log.Printf("Payment job %s permanently failed: %v", job.ID, err)
}
