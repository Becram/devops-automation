package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/Becram/devops-automation/pkg/config"
	"github.com/Becram/devops-automation/pkg/logging"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	enableFeatureFlag = "enable-feature"
	htmlVersion       = `<html>
<head><title>DevOps Automation</title></head>
<body>
<h1>Thanks for using :)</h1>
Version: %s
<p><a href="/metrics">Metrics</a></p>
%s
</body>
</html>`
	htmlPprof = `<p><a href="/debug/pprof">Pprof</a><p>`
)

var version = "custom-build"

var (
	addr             string
	configFile       string
	debug            bool
	profilingEnabled bool

	logger logging.Logger

	cfg = config.ScrapeConf{}
)

func main() {
	app := NewApp()
	if err := app.Run(os.Args); err != nil {
		logger.Error(err, "Error running")
		os.Exit(1)
	}
}

// NewautoApp creates a new cli.App implementing the auto entrypoints and CLI arguments.
func NewApp() *cli.App {
	auto := cli.NewApp()
	auto.Name = "DevOps Automation"
	auto.Version = version
	auto.Usage = "app configured to perform various devops tasks"
	auto.Description = ""
	auto.Authors = []*cli.Author{
		{Name: "", Email: ""},
	}

	auto.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "listen-address",
			Value:       ":8181",
			Usage:       "The address to listen on",
			Destination: &addr,
			EnvVars:     []string{"listen-address"},
		},
		&cli.StringFlag{
			Name:        "config.file",
			Value:       "config.yml",
			Usage:       "Path to configuration file",
			Destination: &configFile,
			EnvVars:     []string{"config.file"},
		},
		&cli.BoolFlag{
			Name:        "debug",
			Value:       false,
			Usage:       "Verbose logging",
			Destination: &debug,
			EnvVars:     []string{"debug"},
		},
		&cli.StringSliceFlag{
			Name:  enableFeatureFlag,
			Usage: "Comma-separated list of enabled features",
		},
	}

	auto.Before = func(ctx *cli.Context) error {
		logger = newLogger(debug)
		return nil
	}

	auto.Commands = []*cli.Command{
		{
			Name: "verify-config", Aliases: []string{"vc"}, Usage: "Loads and attempts to parse config file, then exits. Useful for CI/CD validation",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "config.file", Value: "config.yml", Usage: "Path to configuration file.", Destination: &configFile},
			},
			Action: func(c *cli.Context) error {
				logger.Info("Parsing config")
				if err := cfg.Load(configFile, logger); err != nil {
					logger.Error(err, "Couldn't read config file", "path", configFile)
					os.Exit(1)
				}
				logger.Info("Config file is valid", "path", configFile)
				os.Exit(0)
				return nil
			},
		},
		{
			Name: "version", Aliases: []string{"v"}, Usage: "prints current auto version.",
			Action: func(c *cli.Context) error {
				fmt.Println(version)
				os.Exit(0)
				return nil
			},
		},
	}

	auto.Action = startScraper

	return auto
}

// The function starts a web scraper and sets up a server to handle HTTP requests for metrics and
// health checks.
func startScraper(c *cli.Context) error {
	logger.Info("Parsing config")
	if err := cfg.Load(configFile, logger); err != nil {
		return fmt.Errorf("couldn't read %s: %w", configFile, err)
	}

	logger.Info("auto startup completed", "version", version)

	featureFlags := c.StringSlice(enableFeatureFlag)
	s := NewScraper(featureFlags)
	email := NewMailerConf(&cfg)

	mux := http.NewServeMux()

	if profilingEnabled {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	mux.HandleFunc("/", s.makeHandler())
	mux.HandleFunc("/email", email.mailHandler())

	srv := &http.Server{Addr: addr, Handler: mux}
	return srv.ListenAndServe()
}

func newLogger(debug bool) logging.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetOutput(os.Stdout)

	if debug {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logrus.InfoLevel)
	}

	return logging.NewLogger(l)
}
