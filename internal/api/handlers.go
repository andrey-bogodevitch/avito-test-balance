package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"balance/internal/entity"
	"balance/internal/service"
)

type UserService interface {
	GetBalanceByUserID(id int, currency string) (float64, string, error)
	IncreaseBalance(userID int, amount int) error
	DecreaseBalance(userID int, amount int) error
	TransferMoney(senderID int, recipientID int, amount int) error
	GetOperationsByID(userID int) ([]entity.Operation, error)
}

type UserHandler struct {
	userService UserService
}

func NewHandler(us UserService) *UserHandler {
	return &UserHandler{
		userService: us,
	}
}

func (h *UserHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	currency := r.URL.Query().Get("currency")
	userID := r.URL.Query().Get("user_id")
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		sendJsonError(w, err, http.StatusBadRequest)
		return
	}

	balance, currency, err := h.userService.GetBalanceByUserID(userIDInt, currency)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			sendJsonError(w, fmt.Errorf("%w: id %d", err, userIDInt), http.StatusNotFound)
			return
		}
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}

	type Response struct {
		UserID   int     `json:"user_id"`
		Balance  float64 `json:"balance"`
		Currency string  `json:"currency"`
	}

	resp := Response{
		UserID:   userIDInt,
		Balance:  balance,
		Currency: currency,
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

	if req.Amount < 1 {
		sendJsonError(w, fmt.Errorf("amount must be positive"), http.StatusBadRequest)
		return
	}

	err = h.userService.IncreaseBalance(req.UserID, req.Amount)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			sendJsonError(w, err, http.StatusNotFound)
			return
		}

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

	err = h.userService.DecreaseBalance(req.UserID, req.Amount)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			sendJsonError(w, err, http.StatusNotFound)
			return
		}

		if errors.Is(err, service.ErrWrongInput) {
			sendJsonError(w, err, http.StatusBadRequest)
			return
		}

		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) TransferMoney(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		SenderID    int `json:"sender_id"`
		RecipientID int `json:"recipient_id"`
		Amount      int `json:"amount"`
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendJsonError(w, err, http.StatusBadRequest)
		return
	}

	err = h.userService.TransferMoney(req.SenderID, req.RecipientID, req.Amount)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			sendJsonError(w, err, http.StatusNotFound)
			return
		}
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) GetUserOperations(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		sendJsonError(w, err, http.StatusBadRequest)
		return
	}

	operations, err := h.userService.GetOperationsByID(userIDInt)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			sendJsonError(w, fmt.Errorf("%w: id %d", err, userIDInt), http.StatusNotFound)
			return
		}
		sendJsonError(w, err, http.StatusInternalServerError)
		return
	}

	sendJson(w, operations)
}

var ErrInternal = errors.New("internal error")

type jsonError struct {
	Error string `json:"error"`
}

func sendJsonError(w http.ResponseWriter, err error, code int) {
	log.Println(err)
	if errors.Is(err, service.ErrDatabaseFail) {
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
