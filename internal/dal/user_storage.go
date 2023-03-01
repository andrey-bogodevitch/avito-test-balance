package dal

import "database/sql"

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(dbpool *sql.DB) *UserStorage {
	return &UserStorage{
		db: dbpool,
	}
}

func (s *UserStorage) CreateBalance(userID int, amount int) error {
	query := "INSERT INTO user_balances (user_id, balance) values ($1,$2)"
	_, err := s.db.Exec(query, userID, amount)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) IncreaseBalance(userID int, amount int) error {
	query := "UPDATE user_balances SET balance = balance + $1 where user_id = $2"
	_, err := s.db.Exec(query, amount, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) DecreaseBalance(userID int, amount int) error {
	query := "UPDATE user_balances SET balance = balance - $1 where user_id = $2"
	_, err := s.db.Exec(query, amount, userID)
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
		return 0, err
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
		return err
	}

	err = s.decreaseBalanceTx(tx, senderID, amount)
	if err != nil {
		return err
	}

	_, err = s.getBalanceTx(tx, recipientID)
	if err != nil {
		return err
	}

	err = s.increaseBalanceTx(tx, recipientID, amount)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
