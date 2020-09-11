package main

import (
	"encoding/csv"
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

type Transaction interface {
	GetDate() time.Time
	GetAmount() float32
	GetDescription() string
}

type ChaseTransaction struct {
	TransactionDate string
	PostDate        string
	Description     string
	Category        string
	Type            string
	Amount          float32
}

func (c ChaseTransaction) GetDate() time.Time {
	return c.TransactionDate
}

func (c ChaseTransaction) GetAmount() float32 {
	return c.Amount
}

func (c ChaseTransaction) GetDescription() string {
	return c.Description
}

type CodaTransaction struct {
	TransactionDate string
	Amount          float32
}

func LoadChaseTransactions(inputFilePath string) ([]ChaseTransaction, error) {
	transactions := []Transaction{}
	csvfile, err := os.Open(inputFilePath)
	if err != nil {
		return transactions, err
	}

	r := csv.NewReader(csvfile)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return transactions, err
		}

		log.Print(record)
		log.Print(record[0])
		log.Print(record[1])
	}
	return transactions, nil
}

func AuditFinance(account Account, transactions []Transaction, date time.Time) (bool, error) {
	// fetch records from account, filter by date

	// load input file, pare based on source type

	// sort by date
}
