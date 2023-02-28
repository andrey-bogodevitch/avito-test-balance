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

func (s *UserStorage) GetBalance(userID int) (int, error) {
	query:= "SELECT balance from user_balances where user_id = $1"
	row:=s.db.QueryRow(query, userID)

	var balance int
	err:= row.Scan(&balance)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (s *UserStorage) TransferMoney(FirstUserID, SecondUserID, amount int) error {
	_,err := s.GetBalance(FirstUserID)
	if err != nil {
		return err
	}
	err = s.DecreaseBalance(FirstUserID, amount)
	if err != nil {
		return err
	}

	_, err = s.GetBalance(SecondUserID)
	if err != nil {
		return err
	}
	err = s.IncreaseBalance(SecondUserID, amount)
	if err != nil {
		return err
	}

	return nil
}
