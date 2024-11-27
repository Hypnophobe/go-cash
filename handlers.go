package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type getAddressResponse struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type TransactionRequest struct {
	Pkey    string `json:"pkey"`
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

type submittedBlock struct {
	Block         string `json:"block"`
	PreviousBlock string `json:"prevBlock"`
	Address       string `json:"address"`
	Nonce         string `json:"nonce"`
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func getAddress(w http.ResponseWriter, r *http.Request) {
	address := r.PathValue("address")

	if !validateAddress(address) {
		response := map[string]interface{}{"ok": false, "error": "invalid address"}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	balance := queryAddress(sqliteDatabase, address)

	response := getAddressResponse{
		Address: address,
		Balance: balance,
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"ok": true, "data": response})
}

func getAddresses(w http.ResponseWriter, r *http.Request) {
	addresses, err := queryAddresses(sqliteDatabase)
	if err != nil {
		response := map[string]interface{}{"ok": false, "error": "failed to retrieve addresses"}
		writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := map[string]interface{}{
		"ok":        true,
		"addresses": addresses,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{"ok": false, "error": "invalid request body"}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	senderAddress := generateAddress(req.Pkey)
	senderBalance := queryAddress(sqliteDatabase, senderAddress)

	if !validateAddress(req.Address) {
		response := map[string]interface{}{"ok": false, "error": "invalid address"}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	if senderBalance < req.Amount {
		response := map[string]interface{}{"ok": false, "error": "insufficient funds"}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	updateAddress(sqliteDatabase, senderAddress, senderBalance-req.Amount)
	recipientBalance := queryAddress(sqliteDatabase, req.Address)

	if recipientBalance > 0 {
		updateAddress(sqliteDatabase, req.Address, recipientBalance+req.Amount)
		response := map[string]interface{}{"ok": true}
		writeJSONResponse(w, http.StatusOK, response)
		return
	}

	insertAddress(sqliteDatabase, req.Address, req.Amount)
	insertTransaction(sqliteDatabase, senderAddress, req.Amount, req.Address, int(time.Now().Unix()))

	response := map[string]interface{}{"ok": true}
	writeJSONResponse(w, http.StatusOK, response)
}

func getTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	transaction, err := queryTransaction(sqliteDatabase, id)
	if err != nil {
		response := map[string]interface{}{"ok": false, "error": "internal server error"}
		writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}
	if transaction == nil {
		response := map[string]interface{}{"ok": false, "error": "transaction not found"}
		writeJSONResponse(w, http.StatusNotFound, response)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"ok": true, "transaction": transaction})
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := queryTransactions(sqliteDatabase)
	if err != nil {
		response := map[string]interface{}{"ok": false, "error": "failed to retrieve addresses"}
		writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := map[string]interface{}{
		"ok":           true,
		"transactions": transactions,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func getAddressTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := queryAddressTransactions(sqliteDatabase, r.PathValue("address"))
	if err != nil {
		response := map[string]interface{}{"ok": false, "error": "failed to retrieve addresses"}
		writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := map[string]interface{}{
		"ok":           true,
		"transactions": transactions,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func getBlock(w http.ResponseWriter, r *http.Request) {
	block, err := queryBlock(sqliteDatabase)
	if err != nil {
		response := map[string]interface{}{"ok": false, "error": "internal server error"}
		writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"ok": true, "block": block})
}

func getBlocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := queryBlocks(sqliteDatabase)
	if err != nil {
		response := map[string]interface{}{"ok": false, "error": "failed to retrieve blocks"}
		writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := map[string]interface{}{
		"ok":     true,
		"blocks": blocks,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func submitBlock(w http.ResponseWriter, r *http.Request) {
	var req submittedBlock

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{"ok": false, "error": "invalid request body"}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	qBlock, err := queryBlock(sqliteDatabase)

	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	if qBlock != req.PreviousBlock {
		response := map[string]interface{}{"ok": false, "error": "previous block mismatch"}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	if req.Block != genBlock(req.PreviousBlock, req.Address, req.Nonce) {
		response := map[string]interface{}{"ok": false, "error": "invalid block"}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	insertBlock(sqliteDatabase, req.Block, req.PreviousBlock, req.Address, req.Nonce, int(time.Now().Unix()))
	oldBalance := queryAddress(sqliteDatabase, req.Address)

	if oldBalance > 0 {
		updateAddress(sqliteDatabase, req.Address, oldBalance+1)
		response := map[string]interface{}{"ok": true}
		writeJSONResponse(w, http.StatusOK, response)
		return
	}

	insertAddress(sqliteDatabase, req.Address, oldBalance+1)
	insertTransaction(sqliteDatabase, "null", 1, req.Address, int(time.Now().Unix()))

	response := map[string]interface{}{"ok": true}
	writeJSONResponse(w, http.StatusOK, response)
}

func getTotalSupply(w http.ResponseWriter, r *http.Request) {
	totalBalance, err := getSupply(sqliteDatabase)
	if err != nil {
		response := map[string]interface{}{"ok": false, "error": "internal server error"}
		writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := map[string]interface{}{
		"ok":          true,
		"totalSupply": totalBalance,
	}

	writeJSONResponse(w, http.StatusOK, response)
}
