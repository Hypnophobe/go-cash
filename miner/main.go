package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const serverURL = "http://localhost:8080"

type BlockResponse struct {
	Block string `json:"block"`
	Ok    bool   `json:"ok"`
}

type SubmittedBlock struct {
	Block         string `json:"block"`
	PreviousBlock string `json:"prevBlock"`
	Address       string `json:"address"`
	Nonce         string `json:"nonce"`
}

type Address struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type GetAddressResponse struct {
	Addresses []Address `json:"addresses"`
	Ok        bool      `json:"ok"`
}

func getPrevBlock() (string, error) {
	resp, err := http.Get(serverURL + "/block")
	if err != nil {
		return "", fmt.Errorf("failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	var blockResp BlockResponse
	if err := json.NewDecoder(resp.Body).Decode(&blockResp); err != nil {
		return "", fmt.Errorf("failed to decode GET response: %v", err)
	}

	if !blockResp.Ok {
		return "", fmt.Errorf("GET response was not successful")
	}

	return blockResp.Block, nil
}

func submitBlock(prevBlock string) (bool, error) {
	block := generateBlock(prevBlock)

	subBlock := SubmittedBlock{
		Block:         block,
		PreviousBlock: prevBlock,
		Address:       *address,
		Nonce:         "nonce",
	}

	body, err := json.Marshal(subBlock)
	if err != nil {
		return false, fmt.Errorf("failed to marshal POST body: %v", err)
	}

	resp, err := http.Post(serverURL+"/block", "application/json", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	if len(responseBody) == 0 {
		return false, fmt.Errorf("received empty response body")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return false, fmt.Errorf("failed to decode POST response: %v", err)
	}

	ok, exists := result["ok"].(bool)
	if !exists || !ok {
		return false, fmt.Errorf("POST response not successful")
	}

	return true, nil
}

func generateBlock(prevBlock string) string {
	data := prevBlock + *address + "nonce"
	hash := sha256.New()
	hash.Write([]byte(data))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getBalance(address string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("%s/address/%s", serverURL, address))
	if err != nil {
		return 0, fmt.Errorf("failed to fetch balance: %v", err)
	}
	defer resp.Body.Close()

	var balanceResp GetAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return 0, fmt.Errorf("failed to decode balance response: %v", err)
	}

	if !balanceResp.Ok || len(balanceResp.Addresses) == 0 {
		return 0, fmt.Errorf("could not fetch balance")
	}

	return balanceResp.Addresses[0].Balance, nil
}

var address = flag.String("a", "", "The address to deposit mined funds")

func main() {
	flag.Parse()

	if *address == "" {
		fmt.Println("Error: address is required")
		flag.Usage()
		os.Exit(1)
	}

	for {
		prevBlock, err := getPrevBlock()
		if err != nil {
			log.Fatalf("Error fetching previous block: %v", err)
		}
		fmt.Printf("prevBlock: %s\n", prevBlock)

		newBlock := generateBlock(prevBlock)
		fmt.Printf("newBlock: %s\n", newBlock)

		ok, err := submitBlock(prevBlock)
		if err != nil {
			log.Fatalf("Error submitting block: %v", err)
		}
		if !ok {
			log.Fatal("Block submission failed")
		}

		balance, err := getBalance(*address)
		if err != nil {
			log.Fatalf("Error fetching balance: %v", err)
		}

		fmt.Printf("SUCCESS:%s:%d\n", *address, balance)
	}
}
