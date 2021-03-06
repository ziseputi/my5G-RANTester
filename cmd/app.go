package main

import (
	"my5G-RANTester/config"
	// "fmt"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"my5G-RANTester/internal/templates"
	"os"
)

const version = "0.1"

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
	spew.Config.Indent = "\t"

	log.Info("my5G-RANTester version " + version)

}

func execLoadTest(name string, numberUes int) {
	switch name {
	case "tnla":
		templates.TestMultiAttachUesInConcurrencyWithTNLAs(numberUes)
	case "gnb":
		templates.TestMultiAttachUesInConcurrencyWithGNBs(numberUes)
	default:
		templates.TestMultiAttachUesInQueue(numberUes)
	}
}

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "load-test",
				Aliases: []string{"load-test"},
				Usage: "\nLoad endurance stress tests.\n" +
					"Example for ues in queue: load-test -n 5 \n" +
					"Example for concurrency testing with different GNBs: load-test -g -n 10\n" +
					"Example for concurrency testing with some TNLAs: load-test -t -n 10\n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "number-of-ues", Value: 1, Aliases: []string{"n"}},
					&cli.BoolFlag{Name: "gnb", Aliases: []string{"g"}},
					&cli.BoolFlag{Name: "tnla", Aliases: []string{"t"}},
				},
				Action: func(c *cli.Context) error {
					var execName string
					var name string
					var numUes int
					cfg := config.Data

					if c.IsSet("number-of-ues") {
						execName = "queue"
						name = "Testing multiple UEs attached in queue"
						numUes = c.Int("number-of-ues")
					} else {
						log.Info(c.Command.Usage)
						return nil
					}

					if c.Bool("tnla") {
						execName = "tnla"
						name = "Testing multiple UEs attached in concurrency with TNLAs"
					} else if c.Bool("gnb") {
						execName = "gnb"
						name = "Testing multiple UEs attached in concurrency with different GNBs"
					}
					log.Info("---------------------------------------")
					log.Info("Starting test function: ", name)
					log.Info("Number of UEs: ", numUes)
					log.Info("gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("UPF IP/Port: ", cfg.UPF.Ip, "/", cfg.UPF.Port)
					log.Info("---------------------------------------")
					execLoadTest(execName, numUes)

					return nil
				},
			},
			{
				Name:    "ue",
				Aliases: []string{"ue"},
				Usage:   "Testing an ue attached with configuration",
				Action: func(c *cli.Context) error {
					name := "Testing an ue attached with configuration"
					cfg := config.Data

					log.Info("---------------------------------------")
					log.Info("Starting test function: ", name)
					log.Info("Number of UEs: ", 1)
					log.Info("gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("UPF IP/Port: ", cfg.UPF.Ip, "/", cfg.UPF.Port)
					log.Info("---------------------------------------")
					templates.TestAttachUeWithConfiguration()
					return nil
				},
			},
			{
				Name:    "gnb",
				Aliases: []string{"gnb"},
				Usage: "Testing multiple GNBs attached.\n" +
					"Example for testing attached gnbs: gnb -n 5",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "number-of-gnbs", Value: 1, Aliases: []string{"n"}},
				},
				Action: func(c *cli.Context) error {
					var numGnbs int

					if c.IsSet("number-of-gnbs") {
						numGnbs = c.Int("number-of-gnbs")
					} else {
						log.Info(c.Command.Usage)
						return nil
					}

					name := "Testing multiple GNBs attached"
					cfg := config.Data

					log.Info("---------------------------------------")
					log.Info("Starting test function: ", name)
					log.Info("Number of GNBs: ", numGnbs)
					log.Info("gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("UPF IP/Port: ", cfg.UPF.Ip, "/", cfg.UPF.Port)
					log.Info("---------------------------------------")
					templates.TestMultiAttachGnbInConcurrency(numGnbs)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
