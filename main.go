package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/getlantern/systray"
)

const appSettingsPath = "/Documents/helium-systray.json"

type appSettings struct {
	RefreshMinutes int    `json:"refresh_minutes"`
	AccountAddress string `json:"account_address"`
}

type hotspotMenuItem struct {
	MenuItem *systray.MenuItem
	Status   *systray.MenuItem
	Scale    *systray.MenuItem
	R24H     *systray.MenuItem
	R07D     *systray.MenuItem
	R30D     *systray.MenuItem
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// load config file
	appSettings, err := loadAppSettings(appSettingsPath)
	if err != nil {
		systray.SetTitle("Error loading Helium systray config...")
		time.Sleep(3 * time.Second)
		log.Fatalln(err)
	}

	// set loading status
	systray.SetTitle("Calculating HNT summary...")
	systray.SetTooltip("HNT summary for your Helium hotspots in past 24 hours")

	// setup initial config values
	cfg := newConfig(appSettings)

	// get initial list of hotspots
	hotspotsResp, err := getAccountHotspots(cfg.AccountAddress)
	if err != nil {
		systray.SetTitle("Error fetching hotspots")
		fmt.Println(err)
	}

	// populate hotspot data and menu items
	for _, hs := range hotspotsResp.Data {
		cfg.HsMap[hs.Name] = hs
		cfg.HsMenuItems = append(cfg.HsMenuItems, newHotspotMenuItem(hs.Name))
	}

	// set flag for skipping first refresh
	cfg.SkipHotspotRefresh = true

	// add quit button at the end because order matters
	systray.AddSeparator()
	pref := systray.AddMenuItem("Preferences...", "Adjust preferences")
	displayHNT := pref.AddSubMenuItem("display rewards in HNT", "display hotspot rewards in HNT")
	displayDollars := pref.AddSubMenuItem("display rewards in dollars", "display hotspot rewards in USD")
	editConfig := pref.AddSubMenuItem("Edit config...", "Edit the JSON config")
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	// data refresh routine
	go func() {
		for {
			// clear previous sort/total data
			cfg.ClearPreviousData()

			// get new price
			priceResp, err := getPrice()
			if err != nil {
				systray.SetTitle("Error fetching HNT price")
				fmt.Println(err)
			}
			cfg.Price = priceResp.Data.Price

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
			}

			// get rewards for each hotspot
			for name, hs := range cfg.HsMap {
				// track rewards
				rewardsResp, _ := getHotspotRewards(hs.Address)
				cfg.HsRewards[name] = rewardsResp.Data

				// track sorting order and today's reward
				reward := cfg.RewardOn(name, 0)
				cfg.HsSort = append(cfg.HsSort, sortOrder{Name: name, Reward: reward})
				cfg.Total += reward
			}

			cfg.SortHotspotsByReward()
			cfg.UpdateView()
			time.Sleep(time.Duration(cfg.RefreshMinutes) * time.Minute)
		}
	}()

	// click handling routine
	go func() {
		for {
			select {
			case <-displayHNT.ClickedCh:
				cfg.ConvertToDollars = false
				cfg.UpdateView()
			case <-displayDollars.ClickedCh:
				cfg.ConvertToDollars = true
				cfg.UpdateView()
			case <-editConfig.ClickedCh:

				app := ""
				filepath := ""

				if runtime.GOOS == "windows" {
					app = "explorer"
					filepath = "file:///" + appSettingsFullPath()
				} else {
					app = "open"
					filepath = appSettingsFullPath()
				}

				cmd := exec.Command(app, filepath)
				stdout, err := cmd.Output()

				if err != nil {
						fmt.Println(err.Error())
						return
				}

				fmt.Print(string(stdout))	
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

func appSettingsFullPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "home dir not found"
	}

	return homeDir + appSettingsPath
}

func loadAppSettings(path string) (appSettings, error) {
	var as appSettings

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return as, err
	}

	file, err := os.Open(homeDir + path)
	defer file.Close()
	if err != nil {
		return as, err
	}

	rawSettings, err := ioutil.ReadAll(file)
	if err != nil {
		return as, err
	}

	err = json.Unmarshal(rawSettings, &as)
	if err != nil {
		return as, err
	}

	return as, nil
}

func newHotspotMenuItem(name string) hotspotMenuItem {
	item := systray.AddMenuItem(fmt.Sprintf("Loading %v", name), "")
	return hotspotMenuItem{
		MenuItem: item,
		Status:   item.AddSubMenuItem("Loading...", "Loading data..."),
		Scale:    item.AddSubMenuItem("Loading...", "Loading data..."),
		R24H:     item.AddSubMenuItem("Loading...", "Loading data..."),
		R07D:     item.AddSubMenuItem("Loading...", "Loading data..."),
		R30D:     item.AddSubMenuItem("Loading...", "Loading data..."),
	}
}
