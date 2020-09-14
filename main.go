package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
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
						},
						Action: func(c *cli.Context) error {
							fmt.Println("auditing financial data.")

							if len(c.String("filepath")) < 1 {
								log.Println("No filepath specified for audit source, aborting.")
								return nil
							}

							transactions, err := LoadChaseTransactions(c.String("filepath"))
							if err != nil {
								log.Println("Received error when loading transactions: ", err)
								return nil
							}

							log.Println(transactions)
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
