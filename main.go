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
	"github.com/wontaeyang/helium-systray/icon"
)

const appSettingsPath = "/Documents/helium-systray.json"

type appSettings struct {
	RefreshMinutes   int      `json:"refresh_minutes"`
	AccountAddresses []string `json:"account_addresses"`
	HotspotAddresses []string `json:"hotspot_addresses"`
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
	// Load config file
	appSettings, err := loadAppSettings(appSettingsPath)
	if err != nil {
		handleError(err, "Config error")
	}

	fmt.Printf("app settings loaded: %+v \n", appSettings)

	// Set loading status
	systray.SetIcon(icon.AppIconSmol)
	setAppTitle("Loading summary...")

	// Setup initial config values
	cfg := newConfig(appSettings)
	cfg.FetchAllHotspots()
	cfg.SkipHotspotRefresh = true

	// Setup preferences and quit menu items
	systray.AddSeparator()
	pref := systray.AddMenuItem("Preferences...", "Adjust preferences")
	displayHNT := pref.AddSubMenuItem("display rewards in HNT", "display rewards in HNT")
	displayDollars := pref.AddSubMenuItem("display rewards in USD", "display rewards in USD")
	editConfig := pref.AddSubMenuItem("Edit config...", "Edit the JSON config")
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	// Data refresh routine
	go func() {
		for {
			cfg.ClearPreviousData()
			cfg.GetHNTPrice()
			cfg.RefreshAllHotspots()
			cfg.GetHotspotRewards()
			cfg.SortHotspotsByReward()
			cfg.UpdateView()
			cfg.SkipHotspotRefresh = false
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

func setAppTitle(msg string) {
	systray.SetTitle(msg)
}

func handleSoftError(err error, msg string) {
	systray.SetTitle(msg)
	fmt.Println(err)
}

func handleError(err error, msg string) {
	systray.SetTitle(msg)
	time.Sleep(3 * time.Second)
	log.Fatalln(err)
}
