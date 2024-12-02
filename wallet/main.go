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
)

var syncNode = "http://localhost:8080/"

type Address struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type GetAddressResponse struct {
	Addresses []Address `json:"addresses"`
	OK        bool      `json:"ok"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	OK    bool   `json:"ok"`
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  -b string The address to check the balance of")
	fmt.Println("  -s        Send a transaction")
	fmt.Println("  -p string The password for the private key (for send)")
	fmt.Println("  -a int    The amount to send in the transaction (for send)")
	fmt.Println("  -r string The recipient address to send to (for send)")
}

func main() {
	balanceAddress := flag.String("b", "", "The address to check the balance of")
	send := flag.Bool("s", false, "Send a transaction")
	password := flag.String("p", "", "The password for the private key (for send)")
	amount := flag.Int("a", 0, "The amount to send in the transaction (for send)")
	address := flag.String("r", "", "The address to send to (for send)")

	flag.Usage = usage

	flag.Parse()

	if *balanceAddress != "" {
		balance, err := getBalance(*balanceAddress)
		if err != nil {
			log.Fatalf("Error fetching balance: %v", err)
		}
		fmt.Printf("Address: %s\n", *balanceAddress)
		fmt.Printf("Balance: %d\n", balance)
		return
	}

	if *send {
		if *password == "" || *address == "" || *amount <= 0 {
			fmt.Println("Error: When using -s flag, the flags -p, -t, and -a must be provided.")
			flag.Usage()
			return
		}
		sendTransaction(*password, *address, *amount)
		return
	}

	flag.Usage()
}

func getBalance(address string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("%s/address/%s", syncNode, address))
	if err != nil {
		return 0, fmt.Errorf("failed to fetch balance: %v", err)
	}
	defer resp.Body.Close()

	var balanceResp GetAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return 0, fmt.Errorf("failed to decode balance response: %v", err)
	}

	if !balanceResp.OK || len(balanceResp.Addresses) == 0 {
		return 0, fmt.Errorf("could not fetch balance for address %s", address)
	}

	return balanceResp.Addresses[0].Balance, nil
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
