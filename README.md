# Go Cash
Go Cash is a fictional centralized economy, built for the purpose of learning Go.

## Installation

1. Clone the respository:
    ```bash
    git clone https://github.com/Hypnophobe/go-cash.git && cd go-cash
2. Build the desired binaries:
    ```bash
    go build -o gc-server .
    go build -o gc-wallet wallet/main.go
    go build -o gc-miner miner/main.go
## Usage

### Server

Start the server:
```bash
./gc-server -db (database)
```
To initialize the database (only required on the first run), use the -o flag:
```bash
./gc-server -db (database) -o
```
The -o flag will overwrite the database, so it should only be used when setting up for the first time or if you wish to reset the database.

### Wallet

Run `./gc-wallet` without any flags.

### Miner

Run `./gc-miner` without any flags.