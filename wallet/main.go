package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var syncNode = "http://localhost:8080/"

type Data struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type Response struct {
	Data Data `json:"data"`
	OK   bool `json:"ok"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	OK    bool   `json:"ok"`
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "balance":
			if len(os.Args) == 3 {
				getBalance()
			} else {
				printHelp()
			}
		case "send":
			sendTransaction()
		default:
			printHelp()
		}
	} else {
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Printf("  %s balance (address)\n", os.Args[0])
	fmt.Printf("  %s send (password) (address) (amount)\n", os.Args[0])
}

func getBalance() {
	resp, err := http.Get(syncNode + "address/" + os.Args[2])
	handleError(err, "Failed to connect")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	handleError(err, "Failed to read response body")

	var data Response
	handleError(json.Unmarshal(body, &data), "Failed to unmarshal JSON")

	fmt.Printf("Address: %s / Balance: %d\n", data.Data.Address, data.Data.Balance)
}

func generatePkey(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func sendTransaction() {
	if len(os.Args) != 5 {
		printHelp()
		return
	}

	pkey := generatePkey(os.Args[2])
	address := os.Args[3]

	amount, err := strconv.Atoi(os.Args[4])
	handleError(err, "Invalid amount")

	transaction := map[string]interface{}{
		"pkey":    pkey,
		"address": address,
		"amount":  amount,
	}

	body, err := json.Marshal(transaction)
	handleError(err, "Failed to marshal JSON")

	req, err := http.NewRequest("POST", syncNode+"transaction", bytes.NewBuffer(body))
	handleError(err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	handleError(err, "Failed to send request")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, err := io.ReadAll(resp.Body)
		handleError(err, "Failed to read error response body")

		var errorResponse ErrorResponse
		handleError(json.Unmarshal(errorBody, &errorResponse), "Failed to unmarshal error response")

		fmt.Printf("Transaction failed: %s\n", errorResponse.Error)
		return
	}

	fmt.Println("Transaction sent successfully.")
}

func handleError(err error, message string) {
	if err != nil {
		log.Fatalln(message+":", err)
	}
}
