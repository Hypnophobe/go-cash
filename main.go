package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var overwrite bool
	var dbLocation string

	flag.BoolVar(&overwrite, "o", false, "Overwrite the database")
	flag.StringVar(&dbLocation, "db", "", "Path to the database file")
	flag.Parse()

	if dbLocation == "" {
		fmt.Println("Error: Database file name must be specified using the -db flag.")
		os.Exit(1)
	}

	if overwrite {
		initDatabase(dbLocation)
	}

	loadDatabase(dbLocation)
	defer sqliteDatabase.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /address/{address}", getAddress)                  // Get a single address
	mux.HandleFunc("GET /addresses", getAddresses)                        // Get all addresses
	mux.HandleFunc("POST /transaction", createTransaction)                // Create a transaction
	mux.HandleFunc("GET /transaction/{id}", getTransaction)               // Get single transaction by ID
	mux.HandleFunc("GET /transactions/{address}", getAddressTransactions) // Get all transactions relating to an address
	mux.HandleFunc("GET /transactions", getTransactions)                  // Get all transactions from database
	mux.HandleFunc("POST /block", submitBlock)                            // Submit a block
	mux.HandleFunc("GET /block", getBlock)                                // Get last block
	mux.HandleFunc("GET /blocks", getBlocks)                              // Get all blocks
	mux.HandleFunc("GET /supply", getTotalSupply)                         // Get total currency supply

	log.Println("Server listening to :8080")
	http.ListenAndServe(":8080", mux)
}
