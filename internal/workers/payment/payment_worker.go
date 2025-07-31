package workers

import (
	"jvpayments/internal/queue"
	"jvpayments/internal/services"
	"log"
	"math/rand"
	"time"
)

type PaymentWorker struct {
	queue          *queue.RedisPaymentQueue
	paymentService *services.PaymentService
}

func NewPaymentWorker(queue *queue.RedisPaymentQueue, paymentService *services.PaymentService) *PaymentWorker {
	return &PaymentWorker{
		queue:          queue,
		paymentService: paymentService,
	}

}

func (pw *PaymentWorker) ProcessNextJob() {
	job, err := pw.queue.ConsumePaymentJob()
	if err != nil {
		log.Printf("Error consuming payment job: %v", err)
		return
	}

	log.Printf("Processing payment job: %s", job.ID)

	job.PaymentData.UpdateRequestedAt()

	err = pw.paymentService.ProcessPayment(job.PaymentData)
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

func (pw *PaymentWorker) RequeueJob(job *queue.PaymentJob) error {
	baseDelay := time.Duration(1<<uint(job.RetryCount)) * time.Millisecond

	jitterRange := float64(baseDelay) * 0.5
	jitter := time.Duration(rand.Float64() * jitterRange)

	delay := baseDelay + jitter

	log.Printf("Requeuing job %s with delay: %v (base: %v, jitter: %v)", job.ID, delay, baseDelay, jitter)
	time.Sleep(delay)

	return pw.queue.RequeueJob(job)
}

func (pw *PaymentWorker) HandleFailedJob(job *queue.PaymentJob, err error) {
	// TODO: Implement failed job handling
	// - Log to error queue
	// - Send notification
	// - Store in database for manual review
	log.Printf("Payment job %s permanently failed: %v", job.ID, err)
}

func (pw *PaymentWorker) Start() {
	for {
		pw.ProcessNextJob()
	}
}
