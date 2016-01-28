package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/garyburd/redigo/redis"
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
		cli.IntFlag{
			Name:   "exit-code, e",
			EnvVar: "TATTLE_EXIT_CODE",
			Usage:  "Code that the service failed with",
		},
		cli.StringFlag{
			Name:   "redis-uri, r",
			EnvVar: "TATTLE_REDIS_URI",
			Usage:  "Redis server to tattle to",
		},
		cli.StringFlag{
			Name:   "redis-queue, q",
			EnvVar: "TATTLE_REDIS_QUEUE",
			Usage:  "Redis queue to place the tattle in",
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
	exitCode := context.Int("exit-code")
	serviceName := context.String("service")
	uri := context.String("uri")
	redisURI := context.String("redis-uri")
	redisQueue := context.String("redis-queue")

	if applicationName == "" || exitCode == 0 || serviceName == "" || uri == "" || redisURI == "" || redisQueue == "" {
		cli.ShowAppHelp(context)

		if applicationName == "" {
			color.Red("  Missing required flag --application or TATTLE_APPLICATION")
		}
		if exitCode == 0 {
			color.Red("  Missing required flag --exit-code or TATTLE_EXIT_CODE")
		}
		if serviceName == "" {
			color.Red("  Missing required flag --service or TATTLE_SERVICE")
		}
		if uri == "" {
			color.Red("  Missing required flag --uri or TATTLE_URI")
		}
		if redisURI == "" {
			color.Red("  Missing required flag --redis-uri or TATTLE_REDIS_URI")
		}
		if redisQueue == "" {
			color.Red("  Missing required flag --redis-queue or TATTLE_REDIS_QUEUE")
		}
		os.Exit(1)
	}

	logErrorChannel := make(chan error)
	postErrorChannel := make(chan error)

	go func() {
		logErrorChannel <- logJob(redisURI, redisQueue, applicationName, serviceName, exitCode)
	}()
	go func() {
		postErrorChannel <- postToGovernator(uri, applicationName, serviceName, exitCode)
	}()

	logError := <-logErrorChannel
	postError := <-postErrorChannel

	if logError != nil || postError != nil {
		if logError != nil {
			fmt.Fprintln(os.Stderr, color.RedString("logError: %v", logError.Error()))
		}

		if postError != nil {
			fmt.Fprintln(os.Stderr, color.RedString("postError: %v", postError.Error()))
		}

		os.Exit(1)
	}
}

func postToGovernator(uri, applicationName, serviceName string, exitCode int) error {
	exitCodeStr := fmt.Sprintf("%v", exitCode)
	res, err := http.PostForm(uri, url.Values{
		"applicationName": {applicationName},
		"exitCode":        {exitCodeStr},
		"serviceName":     {serviceName},
	})
	if err != nil {
		return err
	}
	if res.StatusCode != 201 {
		message := fmt.Sprintf("Cancellation failed with code '%v'", res.StatusCode)
		return errors.New(message)
	}
	return nil
}

func logJob(redisURI, redisQueue, applicationName, serviceName string, exitCode int) error {
	redisConn, err := redis.DialURL(redisURI)
	if err != nil {
		return err
	}

	logEntry := NewLogEntry(applicationName, serviceName, exitCode)
	logEntryBytes, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}
	fmt.Printf("%v", string(logEntryBytes))
	_, err = redisConn.Do("LPUSH", redisQueue, logEntryBytes)
	return err
}
