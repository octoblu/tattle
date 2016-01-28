package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

func main() {
	app := cli.NewApp()
	app.Name = "tattle"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "application, a",
			EnvVar: "TATTLE_APPLICATION",
			Usage:  "Name of the application that failed us",
		},
		cli.StringFlag{
			Name:   "exit-code, e",
			EnvVar: "TATTLE_EXIT_CODE",
			Usage:  "Code that the service failed with",
		},
		cli.StringFlag{
			Name:   "service, s",
			EnvVar: "TATTLE_SERVICE",
			Usage:  "Name of the service that is specifically at fault",
		},
		cli.StringFlag{
			Name:   "uri, u",
			EnvVar: "TATTLE_URI",
			Usage:  "Uri to POST to with a cancellation notice",
		},
	}
	app.Run(os.Args)
}

func run(context *cli.Context) {
	applicationName := context.String("application")
	exitCode := context.String("exit-code")
	serviceName := context.String("service")
	uri := context.String("uri")

	if applicationName == "" || exitCode == "" || serviceName == "" || uri == "" {
		cli.ShowAppHelp(context)

		if applicationName == "" {
			color.Red("  Missing required flag --application or TATTLE_APPLICATION")
		}
		if exitCode == "" {
			color.Red("  Missing required flag --exit-code or TATTLE_EXIT_CODE")
		}
		if serviceName == "" {
			color.Red("  Missing required flag --service or TATTLE_SERVICE")
		}
		if uri == "" {
			color.Red("  Missing required flag --uri or TATTLE_URI")
		}
		os.Exit(1)
	}

	res, err := http.PostForm(uri, url.Values{
		"applicationName": {applicationName},
		"exitCode":        {exitCode},
		"serviceName":     {serviceName},
	})
	if err != nil {
		log.Panicf("Failure to post to uri: %v", err.Error())
	}
	if res.StatusCode != 201 {
		log.Panicf("Cancellation failed with code '%v'", res.StatusCode)
	}
}
