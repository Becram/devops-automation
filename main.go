package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"

	"devops-automation/pkg/config"
	"devops-automation/pkg/logging"

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
	fips             bool
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
			Value:       ":5000",
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
		&cli.BoolFlag{
			Name:        "fips",
			Value:       false,
			Usage:       "Use FIPS compliant AWS API endpoints",
			Destination: &fips,
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

func startScraper(c *cli.Context) error {
	logger.Info("Parsing config")
	if err := cfg.Load(configFile, logger); err != nil {
		return fmt.Errorf("Couldn't read %s: %w", configFile, err)
	}

	logger.Info("auto startup completed", "version", version)

	featureFlags := c.StringSlice(enableFeatureFlag)

	s := NewScraper(featureFlags)
	// cache := v1.NewClientCache(cfg, fips, logger)
	// for _, featureFlag := range featureFlags {
	// 	if featureFlag == config.AwsSdkV2 {
	// 		logger.Info("Using aws sdk v2")
	// 		var err error
	// 		// Can't override cache while also creating err
	// 		cache, err = v2.NewCache(cfg, fips, logger)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to construct aws sdk v2 client cache: %w", err)
	// 		}
	// 	}
	// }

	// ctx, cancelRunningScrape := context.WithCancel(context.Background())

	mux := http.NewServeMux()

	if profilingEnabled {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	mux.HandleFunc("/metrics", s.makeHandler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pprofLink := ""
		if profilingEnabled {
			pprofLink = htmlPprof
		}

		_, _ = w.Write([]byte(fmt.Sprintf(htmlVersion, version, pprofLink)))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		logger.Info("Parsing config")
		if err := cfg.Load(configFile, logger); err != nil {
			logger.Error(err, "Couldn't read config file", "path", configFile)
			return
		}

		logger.Info("Reset clients cache")
		// go s.decoupled(ctx, logger, cache)
	})

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
