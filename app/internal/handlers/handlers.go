package handlers

import (
	"bank/internal/database"
	"bank/internal/queueManager"
	"bank/pkg/generator"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func TransferHandler(w http.ResponseWriter, r *http.Request, db database.Database, managerQueue chan<- queueManager.ManagerQueueRequest) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusBadRequest)
		return
	} else {

		bearerToken := strings.Split(authorizationHeader, "Bearer ")
		if len(bearerToken) != 2 {
			http.Error(w, "Invalid Authorization header", http.StatusBadRequest)
			return
		} else {

			token := bearerToken[1]
			clientFrom, err := db.GetClientByToken(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusBadRequest)
				return
			}

			toClientId := r.FormValue("to")
			amount := r.FormValue("amount")
			if toClientId == "" {
				http.Error(w, "Missing to client", http.StatusBadRequest)
				return
			}
			if toClientId == "" {
				http.Error(w, "Missing to client", http.StatusBadRequest)
				return
			}

			toClientIdUint, err := strconv.ParseUint(toClientId, 10, 64)
			if err != nil {
				http.Error(w, "Invalid to client id", http.StatusBadRequest)
				return
			}
			amountFloat, err := strconv.ParseFloat(amount, 64)
			if err != nil {
				http.Error(w, "Invalid amount", http.StatusBadRequest)
				return
			}
			if amountFloat <= 0 {
				http.Error(w, "Invalid amount", http.StatusBadRequest)
				return
			}
			if clientFrom.Balance < amountFloat {
				http.Error(w, "Insufficient funds", http.StatusBadRequest)
				return
			}
			if uint(toClientIdUint) == clientFrom.ID {
				http.Error(w, "Invalid to client id", http.StatusBadRequest)
				return

			}
			clientTo, err := db.GetClientById(uint(toClientIdUint))
			if err != nil {
				http.Error(w, "Invalid to client id", http.StatusBadRequest)
				return
			}

			if managerQueue == nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			} else {

				transactionId := generator.GenerateUniqueID(strconv.Itoa(int(clientFrom.ID)), strconv.Itoa(int(clientTo.ID)), amountFloat)

				err = db.CreateTransaction(database.Transaction{
					ID:         transactionId,
					SenderID:   clientFrom.ID,
					ReceiverID: clientTo.ID,
					Status:     "pending",
					Amount:     amountFloat,
				})
				if err != nil {
					http.Error(w, "Server error", http.StatusInternalServerError)
					return
				}

				select {
				case managerQueue <- queueManager.ManagerQueueRequest{
					ClientId:      clientFrom.ID,
					TransactionId: transactionId,
				}:
					log.Println("Transaction created")
				default:
					err = db.TransactionError(database.Transaction{
						ID:         transactionId,
						SenderID:   clientFrom.ID,
						ReceiverID: clientTo.ID,
						Status:     "error",
						Amount:     amountFloat,
					})
					if err != nil {
						log.Println("Db error:", err)
					}
					http.Error(w, "Server busy try again", http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}

}
