package handlers

import (
	"jvpayments/internal/queue"
	"jvpayments/internal/services"
	"jvpayments/internal/types"
	"log"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
)

type WorkerPool struct {
	NumWorkers int
	Jobs       chan types.PaymentRequest
}

func (wp *WorkerPool) Start(ps *services.PaymentService, paymentQueue *queue.RedisPaymentQueue) {
	for i := 0; i < wp.NumWorkers; i++ {
		go func(workerID int) {
			for job := range wp.Jobs {
				err := ps.ProcessPayment(job)
				if err != nil {
					log.Printf("error processing payment: %v", err)
					queueErr := paymentQueue.PublishPaymentJob(job)
					if queueErr != nil {
						log.Printf("failed to make payment request: %v", queueErr)
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
	wp := WorkerPool{NumWorkers: 350, Jobs: make(chan types.PaymentRequest, 700)}
	wp.Start(paymentService, paymentQueue)
	return &PaymentHandler{
		workerPool: wp,
	}
}

var paymentReqPool = sync.Pool{
	New: func() interface{} {
		return &types.PaymentRequest{}
	},
}

func (ph *PaymentHandler) Payments(ctx *fasthttp.RequestCtx) {
	paymentReq := paymentReqPool.Get().(*types.PaymentRequest)
	defer paymentReqPool.Put(paymentReq)

	*paymentReq = types.PaymentRequest{}

	if err := sonic.Unmarshal(ctx.PostBody(), paymentReq); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		response := map[string]string{"error": "Invalid request body"}
		if jsonData, err := sonic.Marshal(response); err == nil {
			ctx.SetContentType("application/json")
			ctx.SetBody(jsonData)
		}
		return
	}

	ph.workerPool.Jobs <- *paymentReq

	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

// func validatePaymentRequest(req types.PaymentRequest) error {
// 	// TODO: Implement validation logic
// 	return nil
// }
