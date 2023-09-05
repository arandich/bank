package main

import (
	"bank/config"
	"bank/internal/database"
	"bank/internal/handlers"
	"bank/internal/queueManager"
	"bank/internal/worker"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {

	var db database.Database = &database.PostgreSQLDatabase{}
	err := db.Connect()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer func() {
		db.Close()
		log.Println("Closing database")
	}()
	log.Println("Connected to database")

	// Обработка syscall для корректной остановки приложения
	done := make(chan struct{})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			err := db.Ping()
			if err != nil {
				stop <- syscall.SIGTERM
			}
			time.Sleep(time.Second * 5)
		}
	}()
	go func() {

		<-stop
		log.Println("Received stop signal")

		close(done)
	}()

	transactionQueue := make(chan *queueManager.TransactionQueue, 10)
	defer close(transactionQueue)

	wg := sync.WaitGroup{}
	managerQueue := make(chan queueManager.ManagerQueueRequest, 10)

	manager := queueManager.NewManager(managerQueue, done, &wg, transactionQueue)
	wg.Add(2)
	go manager.Start()

	// На случай если у нас упал сервер, то проверяем если зависшие транзакции и обрабатываем их если есть
	go func() {
		transactions, err := db.GetTransactions()
		if err != nil {
			log.Println("Error getting transactions:", err)
			return
		}
		for _, transaction := range transactions {
			managerQueue <- queueManager.ManagerQueueRequest{
				ClientId:      transaction.SenderID,
				TransactionId: transaction.ID,
			}
		}
		return
	}()

	workerNumber := config.GetEnvAsInt("WORKER_NUMBER", 3)

	for i := 0; i < workerNumber; i++ {
		wg.Add(1)
		log.Println("Starting worker", i)
		go worker.NewWorker(db, transactionQueue, done, &wg).Start(i)
	}

	http.HandleFunc("/transfer", func(w http.ResponseWriter, r *http.Request) {
		handlers.TransferHandler(w, r, db, managerQueue)
	})

	// По хорошему, нужно выносить слушателя в отдельный сервис и воркеров тоже в отдельный
	// и связать их между собой через брокер сообщений, но в данном примере, если брокеры закрываются, то и
	// нет смысла держать слушателя
	go func() {
		log.Fatal(http.ListenAndServe(":1323", nil))
	}()

	wg.Wait()

	log.Println("Server stopped")
}
