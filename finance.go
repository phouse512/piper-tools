package main

import (
	"encoding/csv"
	"fmt"
	"github.com/phouse512/go-coda"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	ChaseSource = "Chase"
	VenmoSource = "Venmo"

	FinanceDocId        = "sz-gfMWR-I"
	TransactionsTableId = "grid-53oPnJh4Bt"
	DateColumnId        = "c-4ID4XR1ync"
	DebitColumnId       = "c-FVYpl1uPC1"
	CreditColumnId      = "c-VvaO1RyiJN"
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
	timeObj, err := time.Parse("01/02/2006", c.TransactionDate)
	if err != nil {
		panic(err)
	}

	return timeObj
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
	transactions := []ChaseTransaction{}
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

		if len(record) != 6 {
			log.Print("Skipping because of invalid row count.")
			continue
		}

		val, err := strconv.ParseFloat(record[5], 32)
		if err != nil {
			log.Print("Unable to parse value as float, skipping.")
			continue
		}
		newTransaction := ChaseTransaction{
			TransactionDate: record[0],
			PostDate:        record[1],
			Description:     record[2],
			Category:        record[3],
			Type:            record[4],
			Amount:          float32(val),
		}

		transactions = append(transactions, newTransaction)
	}
	return transactions, nil
}

func AuditFinance(account Account, transactions []Transaction, date time.Time) (bool, error) {
	codaClient := coda.DefaultClient(viper.GetString("coda_api_key"))

	dateStringTz := fmt.Sprintf("%sT22:00:00.000-07:00", date.Format("2006-01-02"))
	rowQuery := coda.ListRowsParameters{
		Query: fmt.Sprintf("%s:\"%s\"", DateColumnId, dateStringTz),
	}
	log.Print(rowQuery)
	rows, err := codaClient.ListTableRows(FinanceDocId, TransactionsTableId, rowQuery)
	if err != nil {
		panic(err)
	}

	for _, row := range rows.Rows {
		log.Print(row.Id)
		log.Print(row.Values[DateColumnId])
	}

	log.Printf("Found %d transactions from coda with date.", len(rows.Rows))

	// fetch records from account, filter by date
	var prunedSrcTransactions []Transaction

	for _, transaction := range transactions {
		if transaction.GetDate() == date {
			prunedSrcTransactions = append(prunedSrcTransactions, transaction)
		}
	}

	log.Printf("Found %d pruned transactions from source.", len(prunedSrcTransactions))
	// load input file, pare based on source type
	//

	for _, src := range prunedSrcTransactions {
		log.Print(src)
	}

	// sort by date
	return false, nil
}
