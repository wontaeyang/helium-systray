# helium-systray
Menu bar app for tracking Helium HNT earnings

![app preview](assets/app.png?raw=true)

### Build and running
Use following  commands to build and run the app in the project folder

```
go build
./helium-systray
```

### Configuration
Helium systray requires a JSON config file at `~/Documents/helium-systray.json`. This can be adjusted by changing `configFileName` variable for a new config location.

```
{
  "refresh_minutes": 15,
  "account_address": "{{ your helium account address here }}"
}
```
