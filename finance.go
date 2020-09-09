package main

import (
	"time"
)

const (
	ChaseSource = "Chase"
	VenmoSource = "Venmo"
)

type Account struct {
	Name   string
	CodaId string
}

type Transaction struct {
	Date        time.Date
	Description string
	Category    string
	Amount      float64
}

type ChaseTransaction struct {
	TransactionDate string
	PostDate        string
	Description     string
	Category        string
	Type            string
	Amount          float32
}

func loadTransactions(sourceType string, inputFilePath string) ([]Transaction, error) {

}

func AuditFinance(account Account, inputFile string, date time.Date) {
	// fetch records from account, filter by date

	// load input file, pare based on source type

	// sort by date
}
