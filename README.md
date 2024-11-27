# Go Cash
Go Cash is a fictional centralized economy, built for the purpose of learning Go.

## Installation

Clone the repository, and build the binary:


    git clone https://github.com/Hypnophobe/go-cash && cd go-cash && go build .

## Usage

Run the server using the following command:

    ./go-cash -db <path_to_your_database_file> [-o]


### Flags

- `-db <path_to_your_database_file>`: Path to the SQLite database file. This flag is required.
- `-o`: Optional flag to overwrite the database if it already exists. If omitted, the application will attempt to use the existing database.