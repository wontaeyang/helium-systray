package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var httpClient = http.Client{
	Timeout: httpTimeout * time.Second,
}

func requestGet(url string, model interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Helium API requires user agent to be set in requests
	req.Header.Set("User-Agent", "Helium-Systray/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.Unmarshal(rawBody, model)
}

func getAccountHotspots(address string) (hotspotsResponse, error) {
	path := fmt.Sprintf("https://api.helium.io/v1/accounts/%s/hotspots", address)
	var resp hotspotsResponse
	err := requestGet(path, &resp)
	return resp, err
}

func getHotspot(address string) (hotspotResponse, error) {
	path := fmt.Sprintf("https://api.helium.io/v1/hotspots/%s", address)
	var resp hotspotResponse
	err := requestGet(path, &resp)
	return resp, err
}

func getHotspotRewards(address string) (rewardsResponse, error) {
	// /rewards/sum?min_time=-60 day&max_time=2021-03-26T06:10:12.251Z&bucket=day
	now := time.Now()
	path := fmt.Sprintf("https://api.helium.io/v1/hotspots/%s/rewards/sum?", address)
	query := url.Values{
		"max_time": {now.Format(time.RFC3339)},
		"min_time": {"-60 day"},
		"bucket":   {"day"},
	}.Encode()

	var resp rewardsResponse
	err := requestGet(path+query, &resp)
	return resp, err
}

func getPrice() (priceResponse, error) {
	path := "https://api.helium.io/v1/oracle/prices/current"
	var resp priceResponse
	err := requestGet(path, &resp)
	return resp, err
}
