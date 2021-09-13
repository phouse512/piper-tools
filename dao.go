package main

import (
	"errors"
)

// Data structure to help access Account objects
type AccountDao struct {
	Accounts   []Account
	AccountMap map[string]Account
}

func InitializeAccountDao() *AccountDao {
	// Creates an AccountDao object with the default array of all accounts.
	accounts, err := GetAllAccounts()
	if err != nil {
		panic(err)
	}

	accountMap := make(map[string]Account)
	for _, account := range accounts {
		accountMap[account.CodaId] = account
	}

	return &AccountDao{
		Accounts:   accounts,
		AccountMap: accountMap,
	}
}

func (a AccountDao) ByCodaId(id string) (Account, error) {
	// Looks for an account by its CodaId, returns Account if exists, otherwise an error
	// @id: string
	account, exists := a.AccountMap[id]
	if !exists {
		return Account{}, errors.New("No account exists for CodaId provided.")
	}

	return account, nil
}

func (a AccountDao) List() ([]Account, error) {
	// Returns a list of Account objects

	return a.Accounts, nil
}
