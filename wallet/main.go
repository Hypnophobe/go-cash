package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	balanceAddress := flag.String("b", "", "The address to check the balance of")
	password := flag.String("p", "", "The password for the private key (for send)")
	address := flag.String("t", "", "The address to send to (for send)")
	amount := flag.Int("a", 0, "The amount to send in the transaction (for send)")

	flag.Parse()

	if *balanceAddress != "" {
		getBalance(*balanceAddress)
		return
	}

	if *password != "" && *address != "" && *amount > 0 {
		sendTransaction(*password, *address, *amount)
		return
	}

	printHelp()
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Printf("  %s -b (address)\n", os.Args[0])
	fmt.Printf("  %s -p (password) -t (address) -a (amount)\n", os.Args[0])
}

func getBalance(address string) {
	resp, err := http.Get(syncNode + "address/" + address)
	if err != nil {
		log.Fatalln("Failed to connect:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Failed to read response body:", err)
	}

	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatalln("Failed to unmarshal JSON:", err)
	}

	fmt.Printf("Address: %s / Balance: %d\n", data.Data.Address, data.Data.Balance)
}

func generatePkey(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func sendTransaction(password, address string, amount int) {
	pkey := generatePkey(password)

	transaction := map[string]interface{}{
		"pkey":    pkey,
		"address": address,
		"amount":  amount,
	}

	body, err := json.Marshal(transaction)
	if err != nil {
		log.Fatalln("Failed to marshal JSON:", err)
	}

	req, err := http.NewRequest("POST", syncNode+"transaction", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln("Failed to create request:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Failed to send request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln("Failed to read error response body:", err)
		}

		var errorResponse ErrorResponse
		err = json.Unmarshal(errorBody, &errorResponse)
		if err != nil {
			log.Fatalln("Failed to unmarshal error response:", err)
		}

		fmt.Printf("Transaction failed: %s\n", errorResponse.Error)
		return
	}

	fmt.Println("Transaction sent successfully.")
}
