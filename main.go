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
	viper.SetConfigName("config2")
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
						Name: "fill",
						Usage: "uses fill mode to generate coda transactions",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name: "filepath",
								Usage: "filepath to the csv.",
								Aliases: []string{"f"},
							},
							&cli.StringFlag{
								Name:    "type",
								Usage:   "source type, (Chase|Ally|Venmo)",
								Aliases: []string{"t"},
							},
							&cli.BoolFlag{
								Name: "commit",
								Usage: "commit and save results",
								Aliases: []string{"c"},
							}
						}
					},
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
								Name:    "startDate",
								Usage:   "date string, %m-%d-%y",
								Aliases: []string{"sd"},
							},
							&cli.StringFlag{
								Name:    "endDate",
								Usage:   "date string, %m-%d-%y",
								Aliases: []string{"ed"},
							},
							&cli.StringFlag{
								Name:    "account",
								Usage:   "acccount name",
								Aliases: []string{"a"},
							},
							&cli.StringFlag{
								Name:    "type",
								Usage:   "source type, (Chase|Ally|Venmo)",
								Aliases: []string{"t"},
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

							startDateObj, err := time.Parse("01-02-06", c.String("startDate"))
							if err != nil {
								log.Println("Received error when parsing date: ", err)
								return err
							}

							endDateObj, err := time.Parse("01-02-06", c.String("endDate"))
							if err != nil {
								log.Println("Received error when parsing date: ", err)
								return err
							}

							result, err := AuditHandler(c.String("type"), c.String("filepath"), c.String("account"), startDateObj, endDateObj)
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
