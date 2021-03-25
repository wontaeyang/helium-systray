package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/getlantern/systray"
)

const configFileName = "/Documents/helium-tray.json"

type config struct {
	RefreshMinutes int    `json:"refresh_minutes"`
	LookbackHours  int    `json:"lookback_hours"`
	AccountAddress string `json:"account_address"`
}

type hotspot struct {
	Name     string
	Status   string
	Address  string
	Reward   float64
	MenuItem *systray.MenuItem
}

type hotspotList map[string]hotspot

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// load config file
	config := loadConfig(configFileName)
	fmt.Printf("Config loaded: %+v", config)

	// set tooltip info about rolling N hours
	systray.SetTitle("Calculating HNT summary...")
	systray.SetTooltip(fmt.Sprintf("HNT summary for rolling %d hours", config.LookbackHours))

	// grab initial account hotspot data
	totalRewards := 0.0
	hotspotList, err := getAccountHotspots(config.AccountAddress)
	if err != nil {
		log.Fatalln(err)
	}

	// add quit button
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	// data refresh routine
	go func() {
		for {
			// get sums for each hotspot and update rows
			for _, hs := range hotspotList {
				reward, _ := getHotspotRewardSum(hs.Address)
				hotspotList[hs.Name].MenuItem.SetTitle(fmt.Sprintf("%s %s - %s", statusToEmoji(hs.Status), floatToString(reward), hs.Name))
				totalRewards += reward
			}

			// update title with total
			systray.SetTitle(fmt.Sprintf("HNT earned: %s", floatToString(totalRewards)))

			// sleep until next refresh
			totalRewards = 0.0
			time.Sleep(time.Duration(config.RefreshMinutes) * time.Minute)
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

func loadConfig(fileName string) config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	configFile, err := os.Open(homeDir + fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer configFile.Close()

	rawConfig, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	var cfg config
	err = json.Unmarshal(rawConfig, &cfg)
	if err != nil {
		log.Fatalln(err)
	}

	return cfg
}

func floatToString(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func statusToEmoji(status string) string {
	if status == "online" {
		return "🟢"
	}
	return "🔴"
}
