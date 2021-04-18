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
make build
```

### Configuration
Helium systray requires a JSON config file at `~/Documents/helium-systray.json`. This can be adjusted by changing `configFileName` variable for a new config location.

```
{
  "refresh_minutes": 15,
  "account_addresses": ["{{ your helium account addresses here }}"],
  "hotspot_addresses": ["{{ individual hotspot addresses here }}"],
}
```

### Credits
Helium Systray icons are designed by [@chadpugh](https://github.com/chadpugh) ( [chadpugh.com](http://chadpugh.com) )
