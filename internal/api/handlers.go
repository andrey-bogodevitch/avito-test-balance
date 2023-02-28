package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	userID:= r.URL.Query().Get("user_id")
	userIDInt, err:= strconv.Atoi(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balance, err:= h.storage.GetBalance(userIDInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Balance: ", balance)

}

func (h *UserHandler) IncreaseBalance(w http.ResponseWriter, r *http.Request) {
	//	ожидаем, что в теле запроса напр придет json следующего вида:
	// {"user_id": 1, "amount": 500}
	type Request struct {
		UserID int `json:"user_id"`
		Amount int `json:"amount"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_,err = h.storage.GetBalance(req.UserID)
	if err != nil {
		err = h.storage.CreateBalance(req.UserID, req.Amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	err = h.storage.IncreaseBalance(req.UserID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // не обязательно, так как если явно не указан код ответа, вернется 200
}

func (h *UserHandler) DecreaseBalance (w http.ResponseWriter, r *http.Request) {
	//	ожидаем, что в теле запроса напр придет json следующего вида:
	// {"user_id": 1, "amount": 500}
	type Request struct {
		UserID int `json:"user_id"`
		Amount int `json:"amount"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.storage.GetBalance(req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.storage.DecreaseBalance(req.UserID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // не обязательно, так как если явно не указан код ответа, вернется 200
}



func (h *UserHandler) TransferMoney (w http.ResponseWriter, r *http.Request) {
	//	ожидаем, что в теле запроса напр придет json следующего вида:
	// {"user1_id": 1, "user2_id: 2, "amount": 500}
	type Request struct {
		FirstUserID int `json:"user1_id"`
		SecondUserID int `json:"user2_id"`
		Amount      int `json:"amount"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.TransferMoney(req.FirstUserID, req.SecondUserID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // не обязательно, так как если явно не указан код ответа, вернется 200
}
