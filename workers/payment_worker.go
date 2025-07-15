package workers

import (
	"jvpayments/queue"
	"jvpayments/services"
	"log"
	"time"
)

// PaymentWorker handles processing of payment jobs from the queue
type PaymentWorker struct {
	queue          *queue.PaymentQueue
	paymentService *services.PaymentService
	stopChan       chan struct{}
	isRunning      bool
}

// NewPaymentWorker creates a new payment worker
func NewPaymentWorker() *PaymentWorker {
	return &PaymentWorker{
		queue:          queue.NewPaymentQueue(),
		paymentService: services.NewPaymentService(),
		stopChan:       make(chan struct{}),
		isRunning:      false,
	}
}

// Start begins processing payment jobs from the queue
func (pw *PaymentWorker) Start() {
	if pw.isRunning {
		log.Println("Payment worker is already running")
		return
	}

	pw.isRunning = true
	log.Println("Payment worker started")

	for {
		select {
		case <-pw.stopChan:
			log.Println("Payment worker stopped")
			return
		default:
			pw.processNextJob()
		}
	}
}

// Stop gracefully stops the payment worker
func (pw *PaymentWorker) Stop() {
	if !pw.isRunning {
		return
	}

	pw.isRunning = false
	close(pw.stopChan)
}

// processNextJob processes the next job from the queue
func (pw *PaymentWorker) processNextJob() {
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
			pw.requeueJob(job)
		} else {
			log.Printf("Payment job %s failed after %d retries", job.ID, job.MaxRetries)
			pw.handleFailedJob(job, err)
		}
		return
	}

	log.Printf("Successfully processed payment job %s: %s", job.ID, paymentResp.ID)
}

// requeueJob puts a job back in the queue for retry
func (pw *PaymentWorker) requeueJob(job *queue.PaymentJob) error {
	// Add delay before requeuing (exponential backoff)
	delay := time.Duration(job.RetryCount) * time.Second
	time.Sleep(delay)

	return pw.queue.RequeueJob(job)
}

// handleFailedJob handles jobs that have failed all retry attempts
func (pw *PaymentWorker) handleFailedJob(job *queue.PaymentJob, err error) {
	// TODO: Implement failed job handling
	// - Log to error queue
	// - Send notification
	// - Store in database for manual review
	log.Printf("Payment job %s permanently failed: %v", job.ID, err)
}
