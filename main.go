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
						Action: func(c *cli.Context) error {
							fmt.Println("auditing financial data.")

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
