package main

import (
	"net/http"
	"os"

	"github.com/Becram/devops-automation/pkg/config"
	"github.com/Becram/devops-automation/pkg/email"
)

type mailer struct {
	CFG config.ScrapeConf
}

func NewMailerConf(cfg *config.ScrapeConf) *mailer { //nolint:revive
	return &mailer{
		CFG: *cfg,
	}
}

func (m *mailer) mailHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("running mail handler", "sending email to")
		if err := cfg.Load(configFile, logger); err != nil {
			logger.Error(err, "Couldn't read config file", "path", configFile)
			os.Exit(1)
		}
		// logger.Info("running /email", "sending email to", m.config.Email.SendTo)
		client := email.NewMailerClient(&m.CFG)
		client.SendEmail()

	}
}
