package queueManager

import (
	"log"
	"sync"
	"time"
)

type Manager struct {
	ManagerQueue         chan ManagerQueueRequest
	TransactionQueueChan chan<- *TransactionQueue
	wg                   *sync.WaitGroup
	done                 chan struct{}
}

type TransactionQueue struct {
	ClientId         uint
	TransactionQueue chan string
	Mu               *sync.Mutex
	InWork           bool
	lastSend         time.Time
}

type ManagerQueueRequest struct {
	ClientId      uint
	TransactionId string
}

func NewManager(managerQueue chan ManagerQueueRequest, done chan struct{}, wg *sync.WaitGroup, transactionQueue chan<- *TransactionQueue) *Manager {
	return &Manager{
		ManagerQueue:         managerQueue,
		wg:                   wg,
		TransactionQueueChan: transactionQueue,
		done:                 done,
	}
}

func (m *Manager) Start() {
	defer m.wg.Done()
	log.Println("Starting manager")

	clientMap := make(map[uint]*TransactionQueue)

	go func(clientMap *map[uint]*TransactionQueue) {
		log.Println("Starting manager queue cleaner")
		defer m.wg.Done()
		for {
			select {
			case <-m.done:
				log.Println("Closing manager queue cleaner")
				return
			default:
				for id, val := range *clientMap {
					if val.lastSend.Add(time.Second*15).Before(time.Now()) && val.InWork == false {
						log.Println("Closing transactionClientQueue", id)
						close(val.TransactionQueue)
						delete(*clientMap, id)
					}
				}
				time.Sleep(time.Second * 5)
			}
		}
	}(&clientMap)

	for {
		select {
		case request := <-m.ManagerQueue:
			value, exists := clientMap[request.ClientId]
			var transactionQueue *TransactionQueue
			if exists {
				transactionQueue = value
			} else {
				transactionQueue = &TransactionQueue{
					ClientId:         request.ClientId,
					TransactionQueue: make(chan string, 3),
					Mu:               &sync.Mutex{},
				}
				clientMap[request.ClientId] = transactionQueue
			}
			transactionQueue.lastSend = time.Now()

			transactionQueue.TransactionQueue <- request.TransactionId

			if transactionQueue.InWork == false {
				transactionQueue.InWork = true
				m.TransactionQueueChan <- transactionQueue
			}

			log.Println("Transaction added to queue: ", request.ClientId)

		case <-m.done:
			log.Println("Closing manager")
			return
		}
	}
}
