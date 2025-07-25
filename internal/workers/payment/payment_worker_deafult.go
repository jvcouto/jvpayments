package workers

import (
	"jvpayments/internal/cache"
	"jvpayments/internal/queue"
	"jvpayments/internal/services"
)

type DefaultWorkerBehavior struct {
	paymentQueue        *queue.RedisPaymentQueue
	paymentService      *services.PaymentService
	paymentCacheService *cache.PaymentCache
}

func NewDefaultWorkerBehavior(paymentQueue *queue.RedisPaymentQueue, paymentService *services.PaymentService, paymentCacheService *cache.PaymentCache) *DefaultWorkerBehavior {
	return &DefaultWorkerBehavior{
		paymentQueue:        paymentQueue,
		paymentService:      paymentService,
		paymentCacheService: paymentCacheService,
	}
}

// func (pw *DefaultWorkerBehavior) ProcessNextJob() {
// 	job, err := pw.queue.ConsumePaymentJob()
// 	if err != nil {
// 		log.Printf("Error consuming payment job: %v", err)
// 		return
// 	}

// 	log.Printf("Processing payment job: %s", job.ID)

// 	job.PaymentData.UpdateRequestedAt()

// 	err = pw.paymentService.ProcessPayment(job.PaymentData, config.CONFIG.PaymentApiUrl)
// 	if err != nil {
// 		log.Printf("Error processing payment job %s: %v", job.ID, err)

// 		if job.RetryCount < job.MaxRetries {
// 			job.RetryCount++
// 			log.Printf("Retrying payment job %s (attempt %d/%d)", job.ID, job.RetryCount, job.MaxRetries)
// 			pw.RequeueJob(job)
// 		} else {
// 			log.Printf("Payment job %s failed after %d retries", job.ID, job.MaxRetries)
// 			pw.HandleFailedJob(job, err)
// 		}
// 		return
// 	}

// 	pw.paymentCacheService.StorePayment(cache.PaymentDefaultKey, job.PaymentData)

// 	log.Printf("Successfully processed payment job %s:", job.ID)
// }

// func (pw *DefaultWorkerBehavior) RequeueJob(job *queue.PaymentJob) error {
// 	delay := time.Duration(job.RetryCount) * time.Second
// 	time.Sleep(delay)

// 	return pw.queue.RequeueJob(job)
// }

// func (pw *DefaultWorkerBehavior) HandleFailedJob(job *queue.PaymentJob, err error) {
// 	// TODO: Implement failed job handling
// 	// - Log to error queue
// 	// - Send notification
// 	// - Store in database for manual review
// 	log.Printf("Payment job %s permanently failed: %v", job.ID, err)
// }
