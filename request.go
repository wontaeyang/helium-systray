package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/getlantern/systray"
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

func getAccountHotspots(address string) (hotspotList, error) {
	list := hotspotList{}
	path := fmt.Sprintf("https://api.helium.io/v1/accounts/%s/hotspots", address)

	var resp accountHotspotsResponse
	err := requestGet(path, &resp)
	if err != nil {
		return list, err
	}

	for _, hs := range resp.Data {
		list[hs.Name] = hotspot{
			Name:     hs.Name,
			Status:   hs.Status.Online,
			Address:  hs.Address,
			MenuItem: systray.AddMenuItem(hs.Name, hs.Status.Online),
		}
	}

	return list, nil
}

func getHotspotRewardSum(address string) (float64, error) {
	now := time.Now()
	path := fmt.Sprintf("https://api.helium.io/v1/hotspots/%s/rewards/sum?", address)
	query := url.Values{
		"min_time": {now.Add(-24 * time.Hour).Format(time.RFC3339)},
		"max_time": {now.Format(time.RFC3339)},
	}.Encode()

	var resp hotspotRewardSummaryResponse
	err := requestGet(path+query, &resp)
	if err != nil {
		return 0, err
	}

	return resp.Data.Total, nil
}
