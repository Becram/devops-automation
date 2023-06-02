package main

import (
	"net/http"
)

type scraper struct {
	featureFlags []string
}

func NewScraper(featureFlags []string) *scraper { //nolint:revive
	return &scraper{
		featureFlags: featureFlags,
	}
}

func (s *scraper) makeHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("running /", "feature flags", s.featureFlags)
	}
}
