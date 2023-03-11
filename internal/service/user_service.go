package service

import (
	"errors"
	"fmt"

	"balance/internal/entity"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrDatabaseFail = errors.New("database failed")
	ErrWrongInput   = errors.New("wrong input")
)

const (
	RUB = "RUB"
	USD = "USD"
)

type CurrencyConverter interface {
	Convert(amount int, baseCurrency string, resultCurrency string) (float64, error)
}

type Storage interface {
	CreateBalance(userID int) error
	IncreaseBalance(userID int, amount int) error
	DecreaseBalance(userID int, amount int) error
	GetBalance(userID int) (int, error)
	TransferMoney(senderID, recipientID, amount int) error
	GetUserOperations(userID int, limit int) ([]entity.Operation, error)
}

type User struct {
	storage   Storage
	converter CurrencyConverter
}

func NewUser(s Storage, c CurrencyConverter) *User {
	return &User{
		storage:   s,
		converter: c,
	}
}

func (u *User) GetBalanceByUserID(id int, currency string) (float64, string, error) {
	balance, err := u.storage.GetBalance(id)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get user %d balance %w", id, err)
	}

	if currency == "" || currency == RUB {
		return float64(balance), RUB, nil
	}

	balanceConverted, err := u.converter.Convert(balance, RUB, currency)
	if err != nil {
		return 0, "", fmt.Errorf("failed to convert from %s to %s", RUB, currency)
	}

	return balanceConverted, currency, nil
}

func (u *User) IncreaseBalance(userID int, amount int) error {
	_, err := u.storage.GetBalance(userID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("failed to get user %d balance %w", userID, err)
		}

		err = u.storage.CreateBalance(userID)
		if err != nil {
			return fmt.Errorf("failed to create balance fo user %d: %w", userID, err)
		}
	}

	err = u.storage.IncreaseBalance(userID, amount)
	if err != nil {
		return fmt.Errorf("failed increase balance")
	}

	return nil
}
func (u *User) DecreaseBalance(userID int, amount int) error {
	balance, err := u.storage.GetBalance(userID)
	if err != nil {
		return fmt.Errorf("failed to get user %d balance: %w", userID, err)
	}

	if balance < amount {
		return fmt.Errorf("%w: not enough money", ErrWrongInput)
	}

	err = u.storage.DecreaseBalance(userID, amount)
	if err != nil {
		return fmt.Errorf("failed decrease balance")
	}

	return nil
}

func (u *User) TransferMoney(senderID int, recipientID int, amount int) error {
	err := u.storage.TransferMoney(senderID, recipientID, amount)
	if err != nil {
		return fmt.Errorf("failed to transfer money: %w", err)
	}

	return nil
}

func (u *User) GetOperationsByID(userID int, limit int) ([]entity.Operation, error) {
	operations, err := u.storage.GetUserOperations(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %d operations: %w", userID, err)
	}

	return operations, nil
}
