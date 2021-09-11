package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/olekukonko/tablewriter"
	"github.com/phouse512/go-coda"
	"github.com/spf13/viper"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	MaxDuration = 100
	AllySource  = "Ally"
	ChaseSource = "Chase"
	VenmoSource = "Venmo"

	FinanceDocId        = "sz-gfMWR-I"
	TransactionsTableId = "table-XeiKO3oHhz"
	AccountsTableId     = "grid-Jzlaq7_uYQ"

	AccountsNameColumnId = "c-mwy2jqwnOQ"
	AccountsTypeColumnId = "c-BjJ_UnF3Al"
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
	GetAmount() float64
	GetDescription() string
}

type VenmoTransaction struct {
	Id       string
	Datetime string
	Type     string
	Status   string
	Note     string
	From     string
	To       string
	Amount   float64
}

func (v VenmoTransaction) GetDate() time.Time {
	timeObj, err := time.Parse("2006-01-02T15:04:05", v.Datetime)
	if err != nil {
		panic(err)
	}

	timeStr := timeObj.Format("2006-01-02")
	dateObj, err := time.Parse("2006-01-02", timeStr)
	if err != nil {
		panic(err)
	}

	return dateObj
}

func (v VenmoTransaction) GetAmount() float64 {
	return v.Amount
}

func (v VenmoTransaction) GetDescription() string {
	return v.Note
}

type AllyTransaction struct {
	Date        string
	Time        string
	Amount      float64
	Type        string
	Description string
}

func (a AllyTransaction) GetDate() time.Time {
	timeObj, err := time.Parse("2006-01-02", a.Date)
	if err != nil {
		panic(err)
	}

	return timeObj
}

func (a AllyTransaction) GetAmount() float64 {
	return a.Amount
}

func (a AllyTransaction) GetDescription() string {
	return a.Description
}

type ChaseTransaction struct {
	TransactionDate string
	PostDate        string
	Description     string
	Category        string
	Type            string
	Amount          float64
}

func (c ChaseTransaction) GetDate() time.Time {
	timeObj, err := time.Parse("01/02/2006", c.TransactionDate)
	if err != nil {
		panic(err)
	}

	return timeObj
}

func (c ChaseTransaction) GetAmount() float64 {
	return c.Amount
}

func (c ChaseTransaction) GetDescription() string {
	return c.Description
}

type CodaTransaction struct {
	Id              string
	TransactionDate time.Time
	Amount          float64
	DebitAccountId  string
	CreditAccountId string
}

func NewCodaTransaction(row coda.Row) *CodaTransaction {
	c := new(CodaTransaction)
	c.Id = row.Id
	c.CreditAccountId = row.Values[CreditColumnId].(map[string]interface{})["rowId"].(string)
	c.DebitAccountId = row.Values[DebitColumnId].(map[string]interface{})["rowId"].(string)
	c.Amount = row.Values[AmountColumnId].(map[string]interface{})["amount"].(float64)

	return c
}

func LoadVenmoTransactions(inputFilePath string) ([]VenmoTransaction, error) {
	transactions := []VenmoTransaction{}
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

		if len(record) != 18 {
			log.Print("Skipping because of invalid row count.")
			continue
		}

		strippedStr := strings.ReplaceAll(record[8], " ", "")
		strippedStr = strings.ReplaceAll(strippedStr, "$", "")
		val, err := strconv.ParseFloat(strippedStr, 64)
		if err != nil {
			log.Print("Unable to parse value as float, skipping.")
			continue
		}

		newTransaction := VenmoTransaction{
			Id:       record[1],
			Datetime: record[2],
			Type:     record[3],
			Status:   record[4],
			Note:     record[5],
			From:     record[6],
			To:       record[7],
			Amount:   val,
		}

		transactions = append(transactions, newTransaction)
	}

	return transactions, nil
}

func LoadAllyTransactions(inputFilePath string) ([]AllyTransaction, error) {
	transactions := []AllyTransaction{}
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

		if len(record) != 5 {
			log.Print("Skipping row because of invalid row count.")
			continue
		}

		val, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Print("Unable to parse value as float, skipping.")
			continue
		}

		newTransaction := AllyTransaction{
			Date:        record[0],
			Time:        record[1],
			Amount:      val,
			Type:        record[3],
			Description: record[4],
		}
		transactions = append(transactions, newTransaction)
	}
	return transactions, nil
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

		val, err := strconv.ParseFloat(record[5], 64)
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
			Amount:          float64(val),
		}

		transactions = append(transactions, newTransaction)
	}
	return transactions, nil
}

func LoadTransactions(sourceType, transactionFilePath string) ([]Transaction, error) {
	// Utility wrapper that returns a list of Transaction objects
	// @sourceType: string, <Ally|Chase|Venmo>
	// @transactionFilePath: string, file path for transaction file
	// returns: List of Transactions or error.
	var transactions []Transaction
	if sourceType == ChaseSource {
		sourceTransactions, err := LoadChaseTransactions(transactionFilePath)
		if err != nil {
			return transactions, err
		}

		// manually convert source transaction objects to generic Transaction to satisfy interface
		transactions = make([]Transaction, len(sourceTransactions))
		for i := range sourceTransactions {
			transactions[i] = sourceTransactions[i]
		}
	} else if sourceType == AllySource {
		sourceTransactions, err := LoadAllyTransactions(transactionFilePath)
		if err != nil {
			return transactions, err
		}

		// manually convert source transaction objects to generic Transaction to satisfy interface
		transactions = make([]Transaction, len(sourceTransactions))
		for i := range sourceTransactions {
			transactions[i] = sourceTransactions[i]
		}

	} else if sourceType == VenmoSource {
		sourceTransactions, err := LoadVenmoTransactions(transactionFilePath)
		if err != nil {
			return transactions, err
		}

		// manually convert source transaction objects to generic Transaction to satisfy interface
		transactions = make([]Transaction, len(sourceTransactions))
		for i := range sourceTransactions {
			transactions[i] = sourceTransactions[i]
		}

	} else {
		log.Printf("Invalid source type provided: %s", sourceType)
		return transactions, errors.New("Invalid source type")
	}

	return transactions, nil
}

func filterSrcRows(date time.Time, rows []Transaction) []Transaction {
	var prunedSrcTransactions []Transaction
	for _, transaction := range rows {
		if transaction.GetDate().Equal(date) {
			prunedSrcTransactions = append(prunedSrcTransactions, transaction)
		}
	}

	return prunedSrcTransactions
}

func filterCodaRows(account Account, date time.Time, rows []CodaTransaction) []CodaTransaction {
	// filter out coda rows by account
	var prunedRows []CodaTransaction
	for _, row := range rows {
		if row.DebitAccountId == account.CodaId || row.CreditAccountId == account.CodaId {
			prunedRows = append(prunedRows, row)
		}
	}

	return prunedRows
}

func fetchCodaRowsRange(startDate, endDate time.Time) []CodaTransaction {
	// fetches all coda transaction rows over a time range.
	var allRows []CodaTransaction
	datePointer := startDate
	for i := 0; i < MaxDuration; i++ {

		codaDayRows := fetchCodaRows(datePointer)
		allRows = append(allRows, codaDayRows...)

		datePointer = datePointer.AddDate(0, 0, 1)

		// check if we've eclipsed the end date
		if datePointer.After(endDate) {
			break
		}
	}

	return allRows
}

func fetchCodaRows(date time.Time) []CodaTransaction {
	// special helper function to fetch transactions by date, due to timezone issues with how dates are initialized
	var totalRows []CodaTransaction
	codaClient := coda.DefaultClient(viper.GetString("coda_api_key"))

	updatedDate := date.AddDate(0, 0, -1)
	dateStringTz1 := fmt.Sprintf("%sT22:00:00.000-07:00", updatedDate.Format("2006-01-02"))
	dateStringTz2 := fmt.Sprintf("%sT22:00:00.000-08:00", updatedDate.Format("2006-01-02"))
	rowQuery1 := coda.ListRowsParameters{
		Query:       fmt.Sprintf("%s:\"%s\"", DateColumnId, dateStringTz1),
		ValueFormat: "rich",
	}
	rowQuery2 := coda.ListRowsParameters{
		Query:       fmt.Sprintf("%s:\"%s\"", DateColumnId, dateStringTz2),
		ValueFormat: "rich",
	}
	var rows1, rows2 coda.ListRowsResponse
	var err1, err2 error

	err := retry.Do(
		func() error {
			rows1, err1 = codaClient.ListTableRows(FinanceDocId, TransactionsTableId, rowQuery1)
			rows2, err2 = codaClient.ListTableRows(FinanceDocId, TransactionsTableId, rowQuery2)
			if err1 != nil {
				return err1
			}

			if err2 != nil {
				return err2
			}

			return nil
		},
	)

	if err != nil {
		log.Printf("Unable to get coda rows even with retry: %v", err)
		panic(err)
	}

	// convert coda.Row objects to CodaTransaction
	for i := range rows1.Rows {
		codaTrans := NewCodaTransaction(rows1.Rows[i])
		totalRows = append(totalRows, *codaTrans)
	}

	for i := range rows2.Rows {
		codaTrans := NewCodaTransaction(rows2.Rows[i])
		totalRows = append(totalRows, *codaTrans)
	}

	return totalRows
}

type CodaTransactionBuilder interface {
	Build(tx Transaction) (CodaTransaction, error)
}

type ManualCodaBuilder struct {
	accountDao *AccountDao
}

func (m ManualCodaBuilder) Build(tx Transaction) (CodaTransaction, error) {
	// Implementation for manual coda transaction builder, displays info and prompts for desired
	//   credit/debit accounts.

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Date", "Amount", "Description"})
	// display data about tx
	table.Append([]string{
		tx.GetDate().Format("Mon Jan _2 15:04:05 2006"),
		fmt.Sprintf("%f", tx.GetAmount()),
		tx.GetDescription(),
	})
	table.Render()

	// prompt for input for which account, store if accurate
	return CodaTransaction{}, nil
}

func BuildCodaTransactions(txs []Transaction, builder CodaTransactionBuilder) ([]CodaTransaction, error) {
	// Builds a list of CodaTransactions based on generic source types. Returns a list or an error
	// @txs: []Transaction, list of generic Transaction objects
	// @builder: struct that implements the transaction builder interface

	// loop through txs, get coda transactions
	var codaTxs []CodaTransaction
	for _, tx := range txs {
		codaTx, err := builder.Build(tx)
		if err != nil {
			return codaTxs, err
		}
		codaTxs = append(codaTxs, codaTx)
	}

	return codaTxs, nil
}

type FillParameters struct {
	SourceType          string `json:"sourceType"`
	TransactionFilePath string `json:"transactionFilePath"`
	Commit              bool   `json:"commit"`
}

func FillHandler(fillParams FillParameters) (bool, error) {
	// FillHandler is responsible for filling in CodaRows based on transaction inputs.
	// @sourceType: string, Chase|Venmo|Ally
	// @transactionFilePath: string, filepath to csv inputs
	// @commit: bool, whether or not results will be committed to Coda or not.
	log.Printf("Commit mode: %b", fillParams.Commit)

	// validate inputs
	if !(fillParams.SourceType == AllySource || fillParams.SourceType == ChaseSource || fillParams.SourceType == VenmoSource) {
		return false, errors.New("Provided SourceType not valid.")
	}

	// load account dao
	accountDao := InitializeAccountDao()

	// load Transactions array from file
	transactions, err := LoadTransactions(fillParams.SourceType, fillParams.TransactionFilePath)
	if err != nil {
		return false, err
	}

	// initialize builder
	builder := ManualCodaBuilder{accountDao: accountDao}

	// get rows using build coda rows func
	codaTxs, err := BuildCodaTransactions(transactions, builder)
	log.Print(codaTxs)
	// commit if necessary

	// return status

	return true, nil
}

func AuditHandler(sourceType, transactionFilePath, accountName string, startDate, endDate time.Time) (bool, error) {
	var transactions []Transaction
	if sourceType == ChaseSource {
		sourceTransactions, err := LoadChaseTransactions(transactionFilePath)
		if err != nil {
			return false, err
		}

		// manually convert source transaction objects to generic Transaction to satisfy interface
		transactions = make([]Transaction, len(sourceTransactions))
		for i := range sourceTransactions {
			transactions[i] = sourceTransactions[i]
		}
	} else if sourceType == AllySource {
		sourceTransactions, err := LoadAllyTransactions(transactionFilePath)
		if err != nil {
			return false, err
		}

		// manually convert source transaction objects to generic Transaction to satisfy interface
		transactions = make([]Transaction, len(sourceTransactions))
		for i := range sourceTransactions {
			transactions[i] = sourceTransactions[i]
		}

	} else if sourceType == VenmoSource {
		sourceTransactions, err := LoadVenmoTransactions(transactionFilePath)
		if err != nil {
			return false, err
		}

		// manually convert source transaction objects to generic Transaction to satisfy interface
		transactions = make([]Transaction, len(sourceTransactions))
		for i := range sourceTransactions {
			transactions[i] = sourceTransactions[i]
		}

	} else {
		log.Printf("Invalid source type provided: %s", sourceType)
		return false, errors.New("Invalid source type")
	}

	// find account with provided name
	account, err := SearchAccount(accountName)
	if err != nil {
		return false, err
	}

	isValid, err := AuditFinanceRange(account, transactions, startDate, endDate)
	if err != nil {
		log.Print("Unable to audit finance range with error: %s", err)
		return false, err
	}

	return isValid, nil
}

func displayResults(resultMap map[time.Time]bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Date", "Correct"})

	dateArr := make([]time.Time, len(resultMap))
	i := 0
	for date, _ := range resultMap {
		// https://stackoverflow.com/a/27848197/2457720
		// according to this, better to use index instead of append, so using i for idx
		dateArr[i] = date
		i++
	}

	sort.Slice(dateArr, func(i, j int) bool { return dateArr[i].Before(dateArr[j]) })
	for _, date := range dateArr {
		validStr := "X"
		if resultMap[date] {
			validStr = "\u2713"
		}
		table.Append([]string{date.Format("01-02-06"), validStr})
	}

	table.Render()
}

func AuditFinanceRange(account Account, transactions []Transaction, startDate time.Time, endDate time.Time) (bool, error) {
	auditResults := make(map[time.Time]bool)
	codaReadMap := make(map[string]bool)

	// load all coda transaction rows
	codaRows := fetchCodaRowsRange(startDate, endDate)
	log.Printf("Found %d total rows from Coda in timerange.", len(codaRows))

	datePointer := startDate
	for i := 0; i < MaxDuration; i++ {
		log.Printf("Running audit on date: %s", datePointer)

		isValid, err := AuditFinance(account, transactions, codaRows, datePointer, codaReadMap)
		if err != nil {
			log.Print("Received error when running audit.", err)
			return false, err
		}

		auditResults[datePointer] = isValid

		datePointer = datePointer.AddDate(0, 0, 1)

		// check if we've eclipsed the end date
		if datePointer.After(endDate) {
			break
		}
	}
	displayResults(auditResults)
	return false, nil
}

func AuditFinance(account Account, transactions []Transaction, codaTransactions []CodaTransaction, date time.Time, codaReadMap map[string]bool) (bool, error) {
	/*
	* AuditFinance is responsible for auditing a single day of transactions, returns a boolean if the day is valid or not

	 */
	prunedSrcTransactions := filterSrcRows(date, transactions)
	log.Printf("Found %d pruned transactions from source.", len(prunedSrcTransactions))

	prunedCodaTransactions := filterCodaRows(account, date, codaTransactions)
	log.Printf("Found %d rows from coda to audit.", len(prunedCodaTransactions))

	isCodaRemaining := make(map[string]bool)
	var missingSrcTransactions []Transaction
	for _, srcTrans := range prunedSrcTransactions {
		// search for for corresponding Coda transaction with same value and date
		found := false
		for _, codaRow := range prunedCodaTransactions {

			if val, isPresent := codaReadMap[codaRow.Id]; isPresent {
				if isPresent && val {
					log.Print("Already looked at coda row, skipping.")
					continue
				}
			}

			codaCreditRefId := codaRow.CreditAccountId
			codaDebitRefId := codaRow.DebitAccountId
			if account.IsCredit && srcTrans.GetAmount() < 0 {
				// this means that this is an expense for a credit account, coda account id
				//   should be in credit column
				if math.Abs(srcTrans.GetAmount()) == codaRow.Amount && account.CodaId == codaCreditRefId {
					isCodaRemaining[codaRow.Id] = true
					found = true
				}
			}

			if account.IsCredit && srcTrans.GetAmount() > 0 {
				// this means that the credit account is getting paid off, coda account id
				//   should be in the debit column

				if float64(srcTrans.GetAmount()) == codaRow.Amount && account.CodaId == codaDebitRefId {
					isCodaRemaining[codaRow.Id] = true
					found = true
				}
			}
			//			log.Print(srcTrans.GetAmount())
			//			log.Print(codaRow.Amount)
			//			log.Print(account.CodaId)
			//			log.Print(codaDebitRefId)
			if !account.IsCredit && srcTrans.GetAmount() > 0 {
				// this means that an asset account is increasing, coda account should be in debit column
				if float64(srcTrans.GetAmount()) == codaRow.Amount && account.CodaId == codaDebitRefId {
					isCodaRemaining[codaRow.Id] = true
					found = true
				}
			}

			if !account.IsCredit && srcTrans.GetAmount() < 0 {
				// this means that an asset account is decreasing, coda account should be in
				//   credit column

				if math.Abs(float64(srcTrans.GetAmount())) == codaRow.Amount && account.CodaId == codaCreditRefId {
					isCodaRemaining[codaRow.Id] = true
					found = true
				}
			}

		}

		if !found {
			missingSrcTransactions = append(missingSrcTransactions, srcTrans)
		}
	}

	log.Printf("Had %d missing src transactions for date: %s.", len(missingSrcTransactions), date)

	if len(missingSrcTransactions) > 0 {
		return false, nil
	}

	return true, nil
}
