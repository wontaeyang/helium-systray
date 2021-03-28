package main

import (
	"fmt"
	"sort"

	"github.com/getlantern/systray"
	"github.com/wontaeyang/helium-systray/icon"
)

type sortOrder struct {
	Name   string
	Reward float64
}

type config struct {
	AccountAddress     string              // hotspot account address
	RefreshMinutes     int                 // view refresh minutes
	Total              float64             // total rewards to be displayed in the menu
	SkipHotspotRefresh bool                // option to skip refresh for initial load
	ConvertToDollars   bool                // convert HNT to dollars
	Price              int                 // dollar conversion value
	HsMap              map[string]hotspot  // map of hotspots
	HsRewards          map[string][]reward // 60 day reward data of hotspots
	HsMenuItems        []hotspotMenuItem   // slice of view rows
	HsSort             []sortOrder         // sorting order
}

func (cfg *config) SortHotspotsByReward() {
	sort.SliceStable(cfg.HsSort, func(a, b int) bool {
		return cfg.HsSort[a].Reward > cfg.HsSort[b].Reward
	})
}

func (cfg *config) RewardOn(name string, day int) float64 {
	return cfg.HsRewards[name][day].Total
}

func (cfg *config) RewardDiff(name string, from int, to int) float64 {
	return cfg.RewardOn(name, from) - cfg.RewardOn(name, to)
}

func (cfg *config) rewardToString(val float64) string {
	var result string
	if cfg.ConvertToDollars {
		dollars := val * (float64(cfg.Price) / 100000000)
		result = fmt.Sprintf("$%.2f", dollars)
	} else {
		result = fmt.Sprintf("%.2f", val)
	}
	return result
}

func (cfg *config) UpdateView() {
	for i, order := range cfg.HsSort {
		onlineStatus := cfg.HsMap[order.Name].Status.Online
		rToday := cfg.RewardOn(order.Name, 0)
		rDiff := cfg.RewardDiff(order.Name, 0, 1)

		// update status of each hotspot row
		setStatus(cfg.HsMenuItems[i].MenuItem, onlineStatus, rDiff)
		// update text of each hotspot row
		cfg.HsMenuItems[i].MenuItem.SetTitle(fmt.Sprintf("%s - %s", cfg.rewardToString(rToday), order.Name))
	}

	// update title with total
	systray.SetTitle(fmt.Sprintf("HNT earned: %s", cfg.rewardToString(cfg.Total)))
}

func (cfg *config) ClearPreviousData() {
	cfg.Total = 0.0
	cfg.SkipHotspotRefresh = false
	cfg.HsSort = []sortOrder{}
}

func newConfig(as appSettings) config {
	return config{
		AccountAddress:   as.AccountAddress,
		RefreshMinutes:   as.RefreshMinutes,
		HsMap:            make(map[string]hotspot),
		HsRewards:        make(map[string][]reward),
		HsMenuItems:      []hotspotMenuItem{},
		HsSort:           []sortOrder{},
		ConvertToDollars: false,
	}
}

func setStatus(mi *systray.MenuItem, status string, diff float64) {
	var currentIcon []byte
	switch {
	case status == "online" && diff == 0:
		currentIcon = icon.StatusPos
	case status == "online" && diff > 0:
		currentIcon = icon.StatusPosUp
	case status == "online" && diff < 0:
		currentIcon = icon.StatusPosDown
	case status != "online" && diff == 0:
		currentIcon = icon.StatusErr
	case status != "online" && diff > 0:
		currentIcon = icon.StatusErrUp
	case status != "online" && diff < 0:
		currentIcon = icon.StatusErrDown
	}
	mi.SetIcon(currentIcon)
}
