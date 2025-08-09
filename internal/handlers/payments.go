package handlers

import (
	"jvpayments/internal/queue"
	"jvpayments/internal/services"
	"jvpayments/internal/types"
	"log"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
)

type WorkerPool struct {
	NumWorkers int
	Jobs       chan []byte
}

func (wp *WorkerPool) Start(ps *services.PaymentService, paymentQueue *queue.RedisPaymentQueue) {
	for i := 0; i < wp.NumWorkers; i++ {
		go func(workerID int) {
			for job := range wp.Jobs {

				var paymentReq types.PaymentRequest

				if err := sonic.Unmarshal(job, &paymentReq); err != nil {
					log.Printf("error: %v", err)
					return
				}

				err := ps.ProcessPayment(paymentReq)
				if err != nil {
					log.Printf("[worker pool] error processing payment: %v", err)
					queueErr := paymentQueue.PublishPaymentJob(paymentReq)
					if queueErr != nil {
						log.Printf("failed to publish payment job: %v", queueErr)
					}
				}
			}
		}(i)
	}
}

type PaymentHandler struct {
	workerPool WorkerPool
}

func NewPaymentHandler(paymentService *services.PaymentService, paymentQueue *queue.RedisPaymentQueue) *PaymentHandler {
	wp := WorkerPool{NumWorkers: 400, Jobs: make(chan []byte, 100000)}
	wp.Start(paymentService, paymentQueue)
	return &PaymentHandler{
		workerPool: wp,
	}
}

func (ph *PaymentHandler) Payments(ctx *fasthttp.RequestCtx) {
	bodyCopy := append([]byte(nil), ctx.PostBody()...)

	ph.workerPool.Jobs <- bodyCopy
}
