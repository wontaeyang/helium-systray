package main

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/getlantern/systray"
	"github.com/wontaeyang/helium-systray/icon"
)

type sortOrder struct {
	Name   string
	Reward float64
}

type config struct {
	AccountAddresses   []string            // hotspot account addresses
	HotspotAddresses   []string            // individual hotspot addresses
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

func (cfg *config) FetchAllHotspots() {
	// Get hotspots from accounts
	for _, addr := range cfg.AccountAddresses {
		hotspotsResp, err := getAccountHotspots(addr)
		if err != nil {
			handleError(err, "Failed to fetch hotspots")
		}

		// Populate hotspot data and menu items
		for _, hs := range hotspotsResp.Data {
			cfg.HsMap[hs.Name] = hs
			cfg.HsMenuItems = append(cfg.HsMenuItems, newHotspotMenuItem(hs.Name))
		}
	}

	// Get individual hotspots by address
	for _, addr := range cfg.HotspotAddresses {
		hotspotResp, err := getHotspot(addr)
		if err != nil {
			handleError(err, "Failed to fetch hotspot")
		}

		// Populate hotspot data and menu items
		hs := hotspotResp.Data
		cfg.HsMap[hs.Name] = hs
		cfg.HsMenuItems = append(cfg.HsMenuItems, newHotspotMenuItem(hs.Name))
	}
}

func (cfg *config) GetHNTPrice() {
	// Get new HNT price for conversion
	priceResp, err := getPrice()
	if err != nil {
		msg := "Failed to get HNT price"
		// Hard error on initial launch
		if cfg.SkipHotspotRefresh {
			handleError(err, msg)
		} else {
			handleSoftError(err, msg)
		}
	}
	cfg.Price = priceResp.Data.Price
}

func (cfg *config) RefreshAllHotspots() {
	if !cfg.SkipHotspotRefresh {
		for _, hs := range cfg.HsMap {
			resp, err := getHotspot(hs.Address)
			if err != nil {
				handleSoftError(err, "Failed to refresh hotspots")
			}
			cfg.HsMap[hs.Name] = resp.Data
			cfg.sleep()
		}
	}
}

func (cfg *config) GetHotspotRewards() {
	// Get rewards for each hotspot
	for name, hs := range cfg.HsMap {
		// Track rewards
		rewardsResp, err := getHotspotRewards(hs.Address)
		if err != nil {
			handleSoftError(err, "Failed to get rewards")
		} else {
			cfg.HsRewards[name] = rewardsResp.Data
		}

		// Track sorting order and today's reward
		reward := cfg.RewardOn(name, 0)
		cfg.HsSort = append(cfg.HsSort, sortOrder{Name: name, Reward: reward})
		cfg.Total += reward
		cfg.sleep()
	}
}

func (cfg *config) SortHotspotsByReward() {
	sort.SliceStable(cfg.HsSort, func(a, b int) bool {
		return cfg.HsSort[a].Reward > cfg.HsSort[b].Reward
	})
}

func (cfg *config) RewardOn(name string, day int) float64 {
	return cfg.HsRewards[name][day].Total
}

func (cfg *config) RewardSum(name string, from int, length int) float64 {
	partial := cfg.HsRewards[name][from:length]
	result := float64(0)
	for _, v := range partial {
		result += v.Total
	}
	return result
}

func (cfg *config) RewardDiff(name string, days int) (current float64, previous float64, diff float64) {
	current = cfg.RewardSum(name, 0, days)
	previous = cfg.RewardSum(name, days, 2*days)
	diff = current - previous
	return current, previous, current - previous
}

func (cfg *config) rewardToString(val float64) string {
	var result string
	if cfg.ConvertToDollars {
		dollars := val * (float64(cfg.Price) / 100000000)
		result = fmt.Sprintf("%s USD", floatToString(dollars))
	} else {
		result = fmt.Sprintf("%s HNT", floatToString(val))
	}
	return result
}

func (cfg *config) UpdateView() {
	for i, order := range cfg.HsSort {
		hs := cfg.HsMap[order.Name]
		onlineStatus := hs.Status.Online
		scale := hs.RewardScale

		// Update status of each hotspot row
		r24H, p24H, d24H := cfg.RewardDiff(order.Name, 1)
		setStatus(cfg.HsMenuItems[i].MenuItem, onlineStatus, d24H)
		cfg.HsMenuItems[i].MenuItem.SetTitle(fmt.Sprintf("%s - %s", cfg.rewardToString(r24H), order.Name))

		// Populate sub-menu
		cfg.HsMenuItems[i].Status.SetTitle(fmt.Sprintf("Status: %s", onlineStatus))
		cfg.HsMenuItems[i].Scale.SetTitle(fmt.Sprintf("Reward scale: %s", floatToString(scale)))

		r24HRow := cfg.HsMenuItems[i].R24H
		setStatus(r24HRow, onlineStatus, d24H)
		r24HRow.SetTitle(fmt.Sprintf("24H - %s %s", cfg.rewardToString(r24H), diffPercent(d24H, p24H)))

		r07dRow := cfg.HsMenuItems[i].R07D
		r07d, p07D, d07D := cfg.RewardDiff(order.Name, 7)
		setStatus(r07dRow, onlineStatus, d07D)
		r07dRow.SetTitle(fmt.Sprintf("07D - %s %s", cfg.rewardToString(r07d), diffPercent(d07D, p07D)))

		r30DRow := cfg.HsMenuItems[i].R30D
		r30D, p30D, d30D := cfg.RewardDiff(order.Name, 30)
		setStatus(r30DRow, onlineStatus, d30D)
		r30DRow.SetTitle(fmt.Sprintf("30D - %s %s", cfg.rewardToString(r30D), diffPercent(d30D, p30D)))

		// Set button for opening hotspot in Helium explorer
		cfg.HsMenuItems[i].Explorer.SetTitle("Open Helium explorer...")
	}

	// update title with total
	systray.SetTitle(cfg.rewardToString(cfg.Total))
}

func (cfg *config) ClearPreviousData() {
	cfg.Total = 0.0
	cfg.HsSort = []sortOrder{}
}

func (cfg *config) sleep() {
	time.Sleep(time.Duration(10*len(cfg.HsMap)) * time.Millisecond)
}

func newConfig(as appSettings) config {
	return config{
		AccountAddresses: as.AccountAddresses,
		HotspotAddresses: as.HotspotAddresses,
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

func floatToString(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func diffPercent(diff float64, prev float64) string {
	percent := (diff / prev) * 100
	var prefix string
	switch {
	case math.IsInf(percent, 0):
		return ""
	case math.IsNaN(percent):
		return ""
	case percent > 0:
		prefix = "/ +"
	default:
		prefix = "/ "
	}

	return fmt.Sprintf("%s%s%%", prefix, floatToString(percent))
}
