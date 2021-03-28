package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"

	"github.com/getlantern/systray"
	"github.com/wontaeyang/helium-systray/icon"
)

const appSettingsPath = "/Documents/helium-systray.json"

type appSettings struct {
	RefreshMinutes int    `json:"refresh_minutes"`
	AccountAddress string `json:"account_address"`
}

type hotspotMenuItem struct {
	MenuItem *systray.MenuItem
}

type sortOrder struct {
	Name   string
	Reward float64
}

type config struct {
	Total              float64             // total rewards to be displayed in the menu
	SkipHotspotRefresh bool                // option to skip refresh for initial load
	ConvertToDollars   bool                // convert HNT to dollars
	Price              int                 // dollar conversion value
	HsMap              map[string]hotspot  // map of hotspots
	HsRewards          map[string][]reward // 60 day reward data of hotspots
	HsMenuItems        []hotspotMenuItem   // slice of view rows
	HsSort             []sortOrder         // sorting order
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// load config file
	appSettings := loadAppSettings(appSettingsPath)
	fmt.Printf("Config loaded: %+v", appSettings)

	// set loading status
	systray.SetTitle("Calculating HNT summary...")
	systray.SetTooltip("HNT summary for your Helium hotspots in past 24 hours")

	// setup initial values
	cfg := newConfig()

	// get initial list of hotspots
	hotspotsResp, err := getAccountHotspots(appSettings.AccountAddress)
	if err != nil {
		systray.SetTitle("Error fetching hotspots")
		fmt.Println(err)
	}

	// populate hotspot data and menu items
	for _, hs := range hotspotsResp.Data {
		cfg.HsMap[hs.Name] = hs
		cfg.HsMenuItems = append(cfg.HsMenuItems, newHotspotMenuItem())
	}

	// set flag for skipping first refresh
	cfg.SkipHotspotRefresh = true

	// add quit button at the end because order matters
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	// data refresh routine
	go func() {
		for {
			// get new price
			if cfg.ConvertToDollars {
				priceResp, err := getPrice()
				if err != nil {
					systray.SetTitle("Error fetching HNT price")
					fmt.Println(err)
				}
				cfg.Price = priceResp.Data.Price
			}

			// update hotspots data
			if !cfg.SkipHotspotRefresh {
				hotspotsResp, err := getAccountHotspots(appSettings.AccountAddress)
				if err != nil {
					systray.SetTitle("Error fetching hotspots")
					fmt.Println(err)
				}

				for _, hs := range hotspotsResp.Data {
					cfg.HsMap[hs.Name] = hs
				}

				// TODO: reconcile menu items here
			}

			// get rewards for each hotspot
			for name, hs := range cfg.HsMap {
				// track rewards
				rewardsResp, _ := getHotspotRewards(hs.Address)
				cfg.HsRewards[name] = rewardsResp.Data

				// track sorting order and today's reward
				reward := cfg.HsRewards[name][0].Total
				cfg.HsSort = append(cfg.HsSort, sortOrder{Name: name, Reward: reward})
				cfg.Total += reward
			}

			// sort the hotspots by rewards
			sort.SliceStable(cfg.HsSort, func(a, b int) bool {
				return cfg.HsSort[a].Reward > cfg.HsSort[b].Reward
			})

			// update menu items for each ordered hotspots
			// hsSort, hsMap, hsRewards, hsMenuItems, price
			for i, order := range cfg.HsSort {
				hsStatus := cfg.HsMap[order.Name].Status.Online
				rToday, rDiff := rewardSummary(cfg.HsRewards[order.Name])
				setStatus(cfg.HsMenuItems[i].MenuItem, hsStatus, rDiff)
				cfg.HsMenuItems[i].MenuItem.SetTitle(fmt.Sprintf("%s - %s", floatToString(rToday), order.Name))
			}

			// update title with total
			systray.SetTitle(fmt.Sprintf("HNT earned: %s", floatToString(cfg.Total)))

			// reset values
			cfg.Total = 0.0
			cfg.SkipHotspotRefresh = false
			cfg.HsSort = []sortOrder{}

			// sleep until next refresh
			time.Sleep(time.Duration(appSettings.RefreshMinutes) * time.Minute)
		}
	}()

	// click handling routine
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// no-op
}

func loadAppSettings(path string) appSettings {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	file, err := os.Open(homeDir + path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	rawSettings, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}

	var as appSettings
	err = json.Unmarshal(rawSettings, &as)
	if err != nil {
		log.Fatalln(err)
	}
	return as
}

func newConfig() config {
	return config{
		HsMap:       make(map[string]hotspot),
		HsRewards:   make(map[string][]reward),
		HsMenuItems: []hotspotMenuItem{},
		HsSort:      []sortOrder{},
	}
}

func newHotspotMenuItem() hotspotMenuItem {
	return hotspotMenuItem{
		MenuItem: systray.AddMenuItem("", ""),
	}
}

func updateHotspotData() {
}

func updateHotspotMenuItems() {
}

func floatToString(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func rewardSummary(summary []reward) (today float64, diff float64) {
	switch len(summary) {
	case 0:
		return 0, 0
	case 1:
		return summary[0].Total, 0
	default:
		diff := summary[0].Total - summary[1].Total
		return summary[0].Total, diff
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
