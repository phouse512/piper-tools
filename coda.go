package main

import (
	"errors"
	"fmt"
	"github.com/phouse512/go-coda"
	"github.com/spf13/viper"
)

func listAccounts(searchVal string) ([]Account, error) {
	// listAccounts is an internal function that lists all the Account objects in the finance
	//   document. Optionally allows for a search value by account name.
	// @searchVal: string, search value, if empty, returns all accounts
	codaClient := coda.DefaultClient(viper.GetString("coda_api_key"))

	var rowQuery coda.ListRowsParameters
	if len(searchVal) > 0 {
		rowQuery = coda.ListRowsParameters{
			Query: fmt.Sprintf("%s:\"%s\"", AccountsNameColumnId, searchVal),
		}
	} else {
		rowQuery = coda.ListRowsParameters{}
	}

	var accounts []Account
	rowResp, err := codaClient.ListTableRows(FinanceDocId, AccountsTableId, rowQuery)
	if err != nil {
		return accounts, err
	}

	for _, accountRow := range rowResp.Rows {
		isCredit := true
		if accountRow.Values[AccountsTypeColumnId] == "Asset" {
			isCredit = false
		}
		account := Account{
			Name:     accountRow.Name,
			CodaId:   accountRow.Id,
			IsCredit: isCredit,
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func GetAllAccounts() ([]Account, error) {
	// High-level func to get all accounts from finance database
	return listAccounts("")
}

func SearchAccount(searchVal string) (Account, error) {
	// SearchAccount is a high level function that is used to find a Coda Account object in the
	//   finance database. If no or multiple accounts are found with a name, an error is thrown.
	// @searchVal: string, search value
	accounts, err := listAccounts(searchVal)
	if err != nil {
		return Account{}, err
	}

	if len(accounts) != 1 {
		return Account{}, errors.New("Unable to find accounts with name.")
	}

	return accounts[0], nil
}
