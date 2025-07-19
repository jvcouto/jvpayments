package workers

import (
	"jvpayments/config"
	"jvpayments/queue"
	"jvpayments/services"
	"log"
	"time"
)

type FallbackWorkerBehavior struct {
	queue          *queue.RedisPaymentQueue
	paymentService *services.PaymentService
}

func NewFallbackWorkerBehavior(queue *queue.RedisPaymentQueue, paymentService *services.PaymentService) *FallbackWorkerBehavior {
	return &FallbackWorkerBehavior{
		queue:          queue,
		paymentService: paymentService,
	}
}

func (pw *FallbackWorkerBehavior) ProcessNextJob() {
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

	log.Printf("Successfully processed payment job %s:", job.ID)
}

func (pw *FallbackWorkerBehavior) RequeueJob(job *queue.PaymentJob) error {
	delay := time.Duration(job.RetryCount) * time.Second
	time.Sleep(delay)

	return pw.queue.RequeueJob(job)
}

func (pw *FallbackWorkerBehavior) HandleFailedJob(job *queue.PaymentJob, err error) {
	// TODO: Implement failed job handling
	// - Log to error queue
	// - Send notification
	// - Store in database for manual review
	log.Printf("Payment job %s permanently failed: %v", job.ID, err)
}
