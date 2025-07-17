package workers

import (
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

// processNextJob processes the next job from the queue
func (pw *FallbackWorkerBehavior) ProcessNextJob() {
	// Consume a job from the queue
	job, err := pw.queue.ConsumePaymentJob()
	if err != nil {
		log.Printf("Error consuming payment job: %v", err)
		time.Sleep(1 * time.Second) // Wait before retrying
		return
	}

	log.Printf("Processing payment job: %s", job.ID)

	// Process the payment
	paymentResp, err := pw.paymentService.ProcessPayment(job.PaymentData)
	if err != nil {
		log.Printf("Error processing payment job %s: %v", job.ID, err)

		// Handle retry logic
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

	log.Printf("Successfully processed payment job %s: %s", job.ID, paymentResp.ID)
}

// requeueJob puts a job back in the queue for retry
func (pw *FallbackWorkerBehavior) RequeueJob(job *queue.PaymentJob) error {
	// Add delay before requeuing (exponential backoff)
	delay := time.Duration(job.RetryCount) * time.Second
	time.Sleep(delay)

	return pw.queue.RequeueJob(job)
}

// handleFailedJob handles jobs that have failed all retry attempts
func (pw *FallbackWorkerBehavior) HandleFailedJob(job *queue.PaymentJob, err error) {
	// TODO: Implement failed job handling
	// - Log to error queue
	// - Send notification
	// - Store in database for manual review
	log.Printf("Payment job %s permanently failed: %v", job.ID, err)
}
