package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"balance/internal/entity"
	"balance/internal/service"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(dbpool *sql.DB) *UserStorage {
	return &UserStorage{
		db: dbpool,
	}
}

func (s *UserStorage) CreateBalance(userID int) error {
	query := "INSERT INTO user_balances (user_id, balance) values ($1,$2)"
	_, err := s.db.Exec(query, userID, 0)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) IncreaseBalance(userID int, amount int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = s.increaseBalanceTx(tx, userID, amount)
	if err != nil {
		return err
	}

	oRecipientID := int64(userID)

	operation := entity.Operation{
		Amount:      int64(amount),
		CreatedAt:   time.Now(),
		Description: "Increase balance",
		SenderID:    &oRecipientID,
	}

	err = s.saveOperation(tx, operation)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) DecreaseBalance(userID int, amount int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = s.decreaseBalanceTx(tx, userID, amount)
	if err != nil {
		return err
	}

	oSenderID := int64(userID)

	operation := entity.Operation{
		Amount:      int64(amount),
		CreatedAt:   time.Now(),
		Description: "Decrease balance",
		SenderID:    &oSenderID,
	}

	err = s.saveOperation(tx, operation)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) increaseBalanceTx(tx *sql.Tx, userID int, amount int) error {
	query := "UPDATE user_balances SET balance = balance + $1 where user_id = $2"
	_, err := tx.Exec(query, amount, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) decreaseBalanceTx(tx *sql.Tx, userID int, amount int) error {
	query := "UPDATE user_balances SET balance = balance - $1 where user_id = $2"
	_, err := tx.Exec(query, amount, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) GetBalance(userID int) (int, error) {
	return s.getBalanceTx(nil, userID)
}

// getBalanceTx returns user balance.
// If tx is empty, the query will send without transaction
func (s *UserStorage) getBalanceTx(tx *sql.Tx, userID int) (int, error) {
	query := "SELECT balance from user_balances where user_id = $1"

	var row *sql.Row
	if tx == nil {
		row = s.db.QueryRow(query, userID)
	} else {
		row = tx.QueryRow(query, userID)
	}

	var balance int
	err := row.Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, service.ErrNotFound
		}
		return 0, fmt.Errorf("%w: %s", service.ErrDatabaseFail, err)
	}

	return balance, nil
}

func (s *UserStorage) TransferMoney(senderID, recipientID, amount int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = s.getBalanceTx(tx, senderID)
	if err != nil {
		return fmt.Errorf("get sender %d balance: %w", senderID, err)
	}

	err = s.decreaseBalanceTx(tx, senderID, amount)
	if err != nil {
		return err
	}

	_, err = s.getBalanceTx(tx, recipientID)
	if err != nil {
		return fmt.Errorf("get recipient %d balance: %w", recipientID, err)
	}

	err = s.increaseBalanceTx(tx, recipientID, amount)
	if err != nil {
		return err
	}

	oSenderID := int64(senderID)
	oRecipientID := int64(recipientID)

	operation := entity.Operation{
		Amount:      int64(amount),
		CreatedAt:   time.Now(),
		Description: "Transfer money",
		SenderID:    &oSenderID,
		RecipientID: &oRecipientID,
	}

	err = s.saveOperation(tx, operation)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) saveOperation(tx *sql.Tx, o entity.Operation) error {
	query := "INSERT INTO operations (amount, created_at, description, sender_id, recipient_id) values ($1,$2,$3,$4,$5)"
	_, err := tx.Exec(query, o.Amount, o.CreatedAt, o.Description, o.SenderID, o.RecipientID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) GetUserOperations(userID int, limit int, page int, sort string) ([]entity.Operation, int, error) {
	offset := page*limit - limit
	query := fmt.Sprintf(
		`SELECT id, amount, created_at, description, sender_id, recipient_id
from operations where sender_id = $1 or recipient_id = $1 
order by %s desc limit %d offset %d`, sort, limit, offset,
	)

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	var operations []entity.Operation

	for rows.Next() {
		var op entity.Operation
		err = rows.Scan(&op.ID, &op.Amount, &op.CreatedAt, &op.Description, &op.SenderID, &op.RecipientID)
		if err != nil {
			return nil, 0, err
		}

		operations = append(operations, op)
	}

	var count int
	query = "SELECT COUNT(*) FROM operations where sender_id = $1 or recipient_id = $1"
	err = s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return operations, count, nil
}
