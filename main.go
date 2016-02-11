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
	"github.com/octoblu/go-logentry/logentry"
)

func main() {
	app := cli.NewApp()
	app.Name = "tattle"
	app.Action = run
	app.Version = VERSION
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "docker-url, d",
			EnvVar: "TATTLE_DOCKER_URL",
			Usage:  "Docker URL of the service",
		},
		cli.StringFlag{
			Name:   "etcd-dir, e",
			EnvVar: "TATTLE_ETCD_DIR",
			Usage:  "Etcd dir of the service",
		},
		cli.IntFlag{
			Name:   "exit-code, x",
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
			Name:   "uri, u",
			EnvVar: "TATTLE_URI",
			Usage:  "Uri to POST to with a cancellation notice",
		},
	}
	app.Run(os.Args)
}

func run(context *cli.Context) {
	dockerURL := context.String("docker-url")
	etcdDir := context.String("etcd-dir")
	exitCode := context.Int("exit-code")
	uri := context.String("uri")
	redisURI := context.String("redis-uri")
	redisQueue := context.String("redis-queue")

	if dockerURL == "" || etcdDir == "" || exitCode == 0 || uri == "" || redisURI == "" || redisQueue == "" {
		cli.ShowAppHelp(context)

		if dockerURL == "" {
			color.Red("  Missing required flag --docker-url or TATTLE_DOCKER_URL")
		}
		if etcdDir == "" {
			color.Red("  Missing required flag --etcd-dir or TATTLE_ETCD_DIR")
		}
		if exitCode == 0 {
			color.Red("  Missing required flag --exit-code or TATTLE_EXIT_CODE")
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
		logErrorChannel <- logJob(redisURI, redisQueue, dockerURL, etcdDir, exitCode)
	}()
	go func() {
		postErrorChannel <- postToGovernator(uri, dockerURL, etcdDir, exitCode)
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

func postToGovernator(uri, dockerURL, etcdDir string, exitCode int) error {
	exitCodeStr := fmt.Sprintf("%v", exitCode)
	res, err := http.PostForm(uri, url.Values{
		"dockerUrl": {dockerURL},
		"exitCode":  {exitCodeStr},
		"etcdDir":   {etcdDir},
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

func logJob(redisURI, redisQueue, dockerURL, etcdDir string, exitCode int) error {
	redisConn, err := redis.DialURL(redisURI)
	if err != nil {
		return err
	}

	logEntry := logentry.New("metric:tattle", "tattle", dockerURL, etcdDir, exitCode, 0)
	logEntryBytes, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}

	_, err = redisConn.Do("LPUSH", redisQueue, logEntryBytes)
	return err
}
