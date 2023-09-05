package database

import (
	"bank/config"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type Client struct {
	ID      uint    `db:"id"`
	Name    string  `db:"name"`
	Token   string  `db:"token"`
	Balance float64 `db:"balance"`
}

type Transaction struct {
	ID         string    `db:"id"`
	SenderID   uint      `db:"sender_id"`
	ReceiverID uint      `db:"receiver_id"`
	Status     string    `db:"status"`
	Amount     float64   `db:"amount"`
	CreatedAt  time.Time `db:"created_at"`
}

type Database interface {
	Connect() error
	Close() error
	GetClientByToken(token string) (Client, error)
	GetClientById(id uint) (Client, error)
	MakeTransfer(transaction Transaction) error
	CreateTransaction(transaction Transaction) error
	GetTransaction(id string) (Transaction, error)
	TransactionError(transaction Transaction) error
	GetTransactions() ([]Transaction, error)
	Ping() error
}

type PostgreSQLDatabase struct {
	conn *sql.DB
}

func (db *PostgreSQLDatabase) GetClientByToken(token string) (Client, error) {
	row := db.conn.QueryRow("SELECT * FROM clients WHERE token = $1", token)
	var client Client
	err := row.Scan(&client.ID, &client.Name, &client.Token, &client.Balance)
	if err != nil {
		return Client{}, err
	}
	return client, nil
}

func (db *PostgreSQLDatabase) GetClientById(id uint) (Client, error) {
	row := db.conn.QueryRow("SELECT * FROM clients WHERE id = $1", id)
	var client Client
	err := row.Scan(&client.ID, &client.Name, &client.Token, &client.Balance)
	if err != nil {
		return Client{}, err
	}
	return client, nil
}

func (db *PostgreSQLDatabase) MakeTransfer(transaction Transaction) error {

	sender, err := db.GetClientById(transaction.SenderID)
	if err != nil {
		return err
	}

	if sender.Balance < transaction.Amount {
		return errors.New("insufficient funds")
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Println("Error rolling back transaction:", rollbackErr)
			}
		}
	}()

	_, err = tx.Exec("UPDATE clients SET balance = balance - $1 WHERE id = $2", transaction.Amount, transaction.SenderID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE clients SET balance = balance + $1 WHERE id = $2", transaction.Amount, transaction.ReceiverID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE transactions SET status = 'completed' WHERE id = $1", transaction.ID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *PostgreSQLDatabase) CreateTransaction(transaction Transaction) error {
	_, err := db.conn.Exec("INSERT INTO transactions (id, sender_id, receiver_id, status, amount) VALUES ($1, $2, $3, $4, $5)",
		transaction.ID,
		transaction.SenderID,
		transaction.ReceiverID,
		transaction.Status,
		transaction.Amount,
	)
	if err != nil {
		return err
	}

	return nil
}

func (db *PostgreSQLDatabase) GetTransaction(id string) (Transaction, error) {
	row := db.conn.QueryRow("SELECT * FROM transactions WHERE id = $1", id)
	var transaction Transaction
	err := row.Scan(&transaction.ID, &transaction.SenderID, &transaction.ReceiverID, &transaction.Status, &transaction.Amount, &transaction.CreatedAt)
	if err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}

func (db *PostgreSQLDatabase) GetTransactions() ([]Transaction, error) {
	rows, err := db.conn.Query("SELECT * FROM transactions where status='pending'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.ID, &transaction.SenderID, &transaction.ReceiverID, &transaction.Status, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (db *PostgreSQLDatabase) TransactionError(transaction Transaction) error {
	_, err := db.conn.Exec("UPDATE transactions SET status = 'error' WHERE id = $1", transaction.ID)
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgreSQLDatabase) Connect() error {

	var host = config.GetEnv("PG_HOST", "")

	var port = config.GetEnvAsInt("PG_PORT", 0)

	var user = config.GetEnv("PG_USER", "")
	var password = config.GetEnv("PG_PASSWORD", "")
	var dbname = config.GetEnv("PG_DBNAME", "")

	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db.conn, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	err = db.conn.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (db *PostgreSQLDatabase) Close() error {
	err := db.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgreSQLDatabase) Ping() error {
	err := db.conn.Ping()
	if err != nil {
		return err
	}
	return nil
}
