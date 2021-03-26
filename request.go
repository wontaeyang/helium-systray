package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func requestGet(url string, model interface{}) error {
	resp, err := http.Get(url)
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
