package main

import (
	"fmt"
	"net/http"
)

type scraper struct {
	featureFlags []string
}

func NewScraper(featureFlags []string) *scraper { //nolint:revive
	fmt.Println("new scrapper running")
	return &scraper{
		featureFlags: featureFlags,
	}
}

func (s *scraper) makeHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("running fine")
	}
}
