# Helium Systray
A menu bar app for tracking Helium hotspot earnings. Powered by [systray](https://github.com/getlantern/systray).

![app preview](assets/app.png?raw=true)

### Env setup
- [Install Go](https://golang.org/doc/install)
- [Install 2goarray](https://github.com/cratonica/2goarray)
```sh
go get github.com/cratonica/2goarray
```

### Build and running
Use following  commands to build and run the app in the project folder

```
// For macOS
make build-mac

// For Windows
make build-win
```

### Requirement
Helium systray for mac will require **macOS 10.15 (Catalina)** and above.

### Configuration
Helium systray requires a JSON config file at `~/Documents/helium-systray.json`. This can be adjusted by changing `configFileName` variable for a new config location. You can add hotspots by account addresses or by individual hotspot addresses.

```
{
  "account_addresses": ["{{ your helium account addresses here }}"],
  "hotspot_addresses": ["{{ individual hotspot addresses here }}"],
}
```

## How to automatically start the app on OS restart
* Go to System Preferences > Users & Groups > Login items tab under your profile.
* Click [+] icon and find Helium Systray app.

### Credits
Helium Systray icons are designed by [@chadpugh](https://github.com/chadpugh) ( [chadpugh.com](http://chadpugh.com) )
