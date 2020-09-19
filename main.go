package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"time"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Print("Unable to read in config.json")
		log.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "finance",
				Aliases: []string{"f"},
				Usage:   "access financial data",
				Subcommands: []*cli.Command{
					{
						Name:  "audit",
						Usage: "audit financial data with coda",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "filepath",
								Usage:   "filepath to the csv.",
								Aliases: []string{"f"},
							},
							&cli.StringFlag{
								Name:    "date",
								Usage:   "date string, %m-%d-%y",
								Aliases: []string{"d"},
							},
							&cli.StringFlag{
								Name:    "account",
								Usage:   "acccount name",
								Aliases: []string{"a"},
							},
						},
						Action: func(c *cli.Context) error {
							fmt.Println("auditing financial data.")

							if len(c.String("filepath")) < 1 {
								log.Println("No filepath specified for audit source, aborting.")
								return nil
							}

							if len(c.String("account")) < 1 {
								log.Print("No account specified, aborting.")
								return nil
							}

							chaseTransactions, err := LoadChaseTransactions(c.String("filepath"))
							if err != nil {
								log.Println("Received error when loading transactions: ", err)
								return nil
							}

							account, err := SearchAccount(c.String("account"))
							if err != nil {
								log.Printf("Unable to find account with name: %s", c.String("account"))
								return nil
							}

							dateObj, err := time.Parse("01-02-06", c.String("date"))
							if err != nil {
								log.Println("Received error when parsing date: ", err)
								return err
							}

							transactions := make([]Transaction, len(chaseTransactions))
							for i := range chaseTransactions {
								transactions[i] = chaseTransactions[i]
							}
							result, err := AuditFinance(account, transactions, dateObj)
							log.Print(err)
							log.Print(result)

							return nil
						},
					},
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
