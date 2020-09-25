package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/phouse512/go-coda"
	"github.com/spf13/viper"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

const (
	ChaseSource = "Chase"
	VenmoSource = "Venmo"

	FinanceDocId        = "sz-gfMWR-I"
	TransactionsTableId = "grid-53oPnJh4Bt"
	AccountsTableId     = "grid-Jzlaq7_uYQ"

	AccountsNameColumnId = "c-mwy2jqwnOQ"
	DateColumnId         = "c-4ID4XR1ync"
	DebitColumnId        = "c-FVYpl1uPC1"
	CreditColumnId       = "c-VvaO1RyiJN"
	AmountColumnId       = "c-I5Fa-AJU-7"
)

type Account struct {
	Name     string
	CodaId   string
	IsCredit bool
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

func filterSrcRows(date time.Time, rows []Transaction) []Transaction {
	var prunedSrcTransactions []Transaction
	for _, transaction := range rows {
		if transaction.GetDate() == date {
			prunedSrcTransactions = append(prunedSrcTransactions, transaction)
		}
	}

	return prunedSrcTransactions
}

func filterCodaRows(account Account, rows []coda.Row) []coda.Row {
	// filter out coda rows by account
	var prunedRows []coda.Row
	for _, row := range rows {

		// convert interface{} to map, and access rowId
		creditRefId := row.Values[CreditColumnId].(map[string]interface{})["rowId"]
		debitRefId := row.Values[DebitColumnId].(map[string]interface{})["rowId"]

		if debitRefId == account.CodaId || creditRefId == account.CodaId {
			prunedRows = append(prunedRows, row)
		}
	}

	return prunedRows
}

func SearchAccount(searchVal string) (Account, error) {
	codaClient := coda.DefaultClient(viper.GetString("coda_api_key"))

	rowQuery := coda.ListRowsParameters{
		Query: fmt.Sprintf("%s:\"%s\"", AccountsNameColumnId, searchVal),
	}
	rowResp, err := codaClient.ListTableRows(FinanceDocId, AccountsTableId, rowQuery)
	if err != nil {
		return Account{}, err
	}

	if len(rowResp.Rows) != 1 {
		return Account{}, errors.New("Unable to find accounts with name.")
	}

	return Account{
		Name:     rowResp.Rows[0].Name,
		CodaId:   rowResp.Rows[0].Id,
		IsCredit: true,
	}, nil
}

func AuditFinance(account Account, transactions []Transaction, date time.Time) (bool, error) {
	codaClient := coda.DefaultClient(viper.GetString("coda_api_key"))

	updatedDate := date.AddDate(0, 0, -1)
	dateStringTz := fmt.Sprintf("%sT22:00:00.000-07:00", updatedDate.Format("2006-01-02"))
	rowQuery := coda.ListRowsParameters{
		Query:       fmt.Sprintf("%s:\"%s\"", DateColumnId, dateStringTz),
		ValueFormat: "rich",
	}
	rows, err := codaClient.ListTableRows(FinanceDocId, TransactionsTableId, rowQuery)
	if err != nil {
		panic(err)
	}

	prunedSrcTransactions := filterSrcRows(date, transactions)
	log.Printf("Found %d pruned transactions from source.", len(prunedSrcTransactions))

	prunedCodaTransactions := filterCodaRows(account, rows.Rows)
	log.Printf("Found %d rows from coda to audit.", len(prunedCodaTransactions))

	var isCodaRemaining map[string]bool
	var missingSrcTransactions []Transaction
	for _, srcTrans := range prunedSrcTransactions {
		// search for for corresponding Coda transaction with same value and date
		found := false
		for _, codaRow := range prunedCodaTransactions {

			if val, isPresent := isCodaRemaining[codaRow.Id]; isPresent {
				if isPresent && val {
					log.Print("Already looked at coda row, skipping.")
					continue
				}
			}

			codaVal := codaRow.Values[AmountColumnId].(map[string]interface{})["amount"]
			codaCreditRefId := codaRow.Values[CreditColumnId].(map[string]interface{})["rowId"]
			codaDebitRefId := codaRow.Values[DebitColumnId].(map[string]interface{})["rowId"]
			log.Print(codaVal)
			log.Print(codaCreditRefId)
			if account.IsCredit && srcTrans.GetAmount() < 0 {
				// this means that this is an expense for a credit account, coda account id
				//   should be in credit column

				if math.Abs(float64(srcTrans.GetAmount())) == codaVal && account.CodaId == codaCreditRefId {
					isCodaRemaining[codaRow.Id] = true
					found = true
				}
			}

			if account.IsCredit && srcTrans.GetAmount() > 0 {
				// this means that the credit account is getting paid off, coda account id
				//   should be in the debit column

				if srcTrans.GetAmount() == codaVal && account.CodaId == codaDebitRefId {
					isCodaRemaining[codaRow.Id] = true
					found = true
				}
			}

			if !account.IsCredit && srcTrans.GetAmount() > 0 {
				// this means that an asset account is increasing, coda account should be in debit column
				if srcTrans.GetAmount() == codaVal && account.CodaId == codaDebitRefId {
					isCodaRemaining[codaRow.Id] = true
					found = true
				}
			}

			if !account.IsCredit && srcTrans.GetAmount() < 0 {
				// this means that an asset account is decreasing, coda account should be in
				//   credit column

				if srcTrans.GetAmount() == codaVal && account.CodaId == codaCreditRefId {
					isCodaRemaining[codaRow.Id] = true
					found = true
				}
			}

		}

		if !found {
			missingSrcTransactions = append(missingSrcTransactions, srcTrans)
		}
	}

	log.Printf("Had %d missing src transactions.", len(missingSrcTransactions))

	// sort by date
	return false, nil
}
