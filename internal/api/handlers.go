package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"balance/internal/dal"
)

type Storage interface {
	CreateBalance(userID int, amount int) error
	IncreaseBalance(userID int, amount int) error
	DecreaseBalance(userID int, amount int) error
	GetBalance(userID int) (int, error)
	TransferMoney(userID1, userID2, amount int) error
}

type UserHandler struct {
	storage Storage
}

func NewHandler(s Storage) *UserHandler {
	return &UserHandler{
		storage: s,
	}
}

func (h *UserHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		sendJsonError(w, err, http.StatusBadRequest)
		return
	}

	balance, err := h.storage.GetBalance(userIDInt)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			sendJsonError(w, fmt.Errorf("%w: id %d", err, userIDInt), http.StatusNotFound)
			return
		}
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}

	type Response struct {
		UserID  int `json:"user_id"`
		Balance int `json:"balance"`
	}

	resp := Response{
		UserID:  userIDInt,
		Balance: balance,
	}

	sendJson(w, resp)
}

func (h *UserHandler) IncreaseBalance(w http.ResponseWriter, r *http.Request) {
	//	ожидаем, что в теле запроса напр придет json следующего вида:
	// {"user_id": 1, "amount": 500}
	type Request struct {
		UserID int `json:"user_id"`
		Amount int `json:"amount"`
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendJsonError(w, err, http.StatusBadRequest)
		return
	}

	_, err = h.storage.GetBalance(req.UserID)
	if err != nil {
		if !errors.Is(err, dal.ErrNotFound) {
			sendJsonError(w, err, http.StatusInternalServerError)
			return
		}

		err = h.storage.CreateBalance(req.UserID, req.Amount)
		if err != nil {
			sendJsonError(w, err, http.StatusInternalServerError)
			return
		}
		return
	}

	err = h.storage.IncreaseBalance(req.UserID, req.Amount)
	if err != nil {
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) DecreaseBalance(w http.ResponseWriter, r *http.Request) {
	//	ожидаем, что в теле запроса напр придет json следующего вида:
	// {"user_id": 1, "amount": 500}
	type Request struct {
		UserID int `json:"user_id"`
		Amount int `json:"amount"`
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendJsonError(w, err, http.StatusBadRequest)
		return
	}

	balance, err := h.storage.GetBalance(req.UserID)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			sendJsonError(w, fmt.Errorf("%w: id %d", err, req.UserID), http.StatusNotFound)
			return
		}
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}

	if balance < req.Amount {
		sendJsonError(w, fmt.Errorf("not enough money"), http.StatusBadRequest)
		return
	}

	err = h.storage.DecreaseBalance(req.UserID, req.Amount)
	if err != nil {
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) TransferMoney(w http.ResponseWriter, r *http.Request) {
	//	ожидаем, что в теле запроса напр придет json следующего вида:
	// {"user1_id": 1, "user2_id: 2, "amount": 500}
	type Request struct {
		FirstUserID  int `json:"user1_id"`
		SecondUserID int `json:"user2_id"`
		Amount       int `json:"amount"`
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendJsonError(w, err, http.StatusBadRequest)
		return
	}

	err = h.storage.TransferMoney(req.FirstUserID, req.SecondUserID, req.Amount)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			sendJsonError(w, err, http.StatusNotFound)
			return
		}
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}
}

var ErrInternal = errors.New("internal error")

type jsonError struct {
	Error string `json:"error"`
}

func sendJsonError(w http.ResponseWriter, err error, code int) {
	log.Println(err)
	if errors.Is(err, dal.ErrDatabaseFail) {
		err = ErrInternal
	}
	sendJson(w, jsonError{Error: err.Error()}, code)
}

func sendJson(w http.ResponseWriter, data any, code ...int) {
	w.Header().Set("Content-Type", "application/json")

	if len(code) > 0 {
		w.WriteHeader(code[0])
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}
}
