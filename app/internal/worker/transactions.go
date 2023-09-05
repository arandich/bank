package worker

import (
	"bank/internal/database"
	"bank/internal/queueManager"
	"log"
	"sync"
	"time"
)

type Worker struct {
	Database         database.Database
	transactionQueue <-chan *queueManager.TransactionQueue
	done             <-chan struct{}
	wg               *sync.WaitGroup
}

func NewWorker(database database.Database, transactionQueue <-chan *queueManager.TransactionQueue, done <-chan struct{}, wg *sync.WaitGroup) *Worker {
	return &Worker{
		Database:         database,
		transactionQueue: transactionQueue,
		done:             done,
		wg:               wg,
	}
}

func (w *Worker) Start(workerId int) {
	defer w.wg.Done()

	for {
		select {
		case transactionClientQueue := <-w.transactionQueue:

			log.Println("Worker", workerId, "started work on queue", transactionClientQueue.ClientId)

		loop:
			for {
				select {
				case transaction := <-transactionClientQueue.TransactionQueue:
					transactionDb, err := w.Database.GetTransaction(transaction)
					if err != nil {
						log.Println("Error getting transaction:", err)
						continue
					}
					if transactionDb.Status != "pending" {
						continue
					}
					transactionClientQueue.Mu.Lock()
					// long work
					time.Sleep(time.Second * 5)

					err = w.Database.MakeTransfer(transactionDb)
					if err != nil {
						log.Println("Error making transfer:", err)
						transactionDb.Status = "error"
						err = w.Database.TransactionError(transactionDb)
						if err != nil {
							log.Println("Db error:", err)
						}
						transactionClientQueue.Mu.Unlock()
						continue
					}
					transactionClientQueue.Mu.Unlock()
					log.Println("Transaction", transaction, "completed")
				default:
					break loop
				}
			}
			transactionClientQueue.InWork = false
		case <-w.done:
			log.Println("Worker", workerId, "stopping")
			return
		}
	}

}
