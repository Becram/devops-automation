package config

import (
	"errors"
	"fmt"
	"os"

	"devops-automation/pkg/logging"

	"gopkg.in/yaml.v2"
	// "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/logging"
	// "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
)

type ScrapeConf struct {
	APIVersion string `yaml:"apiVersion"`
	StsRegion  string `yaml:"region"`
}

func (c *ScrapeConf) Load(file string, logger logging.Logger) error {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return err
	}

	logConfigErrors(yamlFile, logger)

	err = c.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (c *ScrapeConf) Validate() error {
	if c.APIVersion != "" && c.APIVersion != "v1alpha1" {
		return fmt.Errorf("unknown apiVersion value '%s'", c.APIVersion)
	}

	return nil
}

// logConfigErrors logs as warning any config unmarshalling error.
func logConfigErrors(cfg []byte, logger logging.Logger) {
	var sc ScrapeConf
	var errMsgs []string
	if err := yaml.UnmarshalStrict(cfg, &sc); err != nil {
		terr := &yaml.TypeError{}
		if errors.As(err, &terr) {
			errMsgs = append(errMsgs, terr.Errors...)
		} else {
			errMsgs = append(errMsgs, err.Error())
		}
	}

	if sc.APIVersion == "" {
		errMsgs = append(errMsgs, "missing apiVersion")
	}

	if len(errMsgs) > 0 {
		for _, msg := range errMsgs {
			logger.Warn("config file syntax error", "err", msg)
		}
		logger.Warn(`Config file error(s) detected: Yace might not work as expected. Future versions of Yace might fail to run with an invalid config file.`)
	}
}
