# helium-systray
Menu bar app for tracking Helium HNT earnings

### Build and running
Use following  commands to build and run the app in the project folder

```
go build
./helium-systray
```

### Configuration
Helium systray requires a JSON config file at `~/Documents/helium-systray.json`. This can be adjusted by chaning `configFileName` variable for a new config location.

```
{
  "refresh_minutes": 10,
  "lookback_hours": 24,
  "account_address": "{{ your helium account address here }}"
}
```
