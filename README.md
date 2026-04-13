# Icinga Dashboard

A very simple dashboard, based on Icinga2, that highlights current host and service issues. Designed to be shown on (for example) Raspberry PI based monitors.

![screenshot](https://github.com/hujiko/icinga-dashboard/raw/master/docs/screenshot.png)

## Setup

You can either compile it manually, or use the prebuilt docker image:

```yaml
services:
  icinga2dashboard:
    image: hujiko/icinga-dashboard:latest
    environment:
      - LISTEN_ADDRESS=":8080"
      - ICINGA2_BASE_URL="https://icinga2.example.com/icingadb"
      - ICINGA2_API_URL="https://icinga2.example.com:5665"
      - ICINGA2_API_USERNAME="admin"
      - ICINGA2_API_PASSWORD="changeme"
    ports:
      - "8080:8080"
```

## Configuration

The dashboard can be configured through environment variables:

```bash
  # Which port to bind to - defaults to 8080
  export LISTEN_ADDRESS=":8080"

  # Hostname of the Icinga2 web interface.
  # When clicking on the dashboard, this is what you will be directed to
  export ICINGA2_BASE_URL="https://icinga2.example.com/icingadb"

  # Hostname of the Icinga2 REST API (with protocol, but without trailing slash)
  export ICINGA2_API_URL="https://icinga2.example.com:5665"

  # Timeout in seconds when calling the icinga2 API
  export ICINGA2_API_TIMEOUT=5

  # If you use certificate based authentication to connect to the Icinga2-API, set those variables
  export ICINGA2_API_CLIENT_KEY_PATH=""
  export ICINGA2_API_CLIENT_CERT_PATH=""

  # If you use a self-signed certificate, point to the CA here for validation
  export ICINGA2_API_CA_PATH=""

  # In case you use a self-signed SSL Certificate for the Icinga2-API
  # but don't want to validate certificates at all, set this.
  # Possible values
  # 0 => Do not validate
  # 1 => Validate certificate
  export ICINGA2_API_VALIDATE_CERTIFICATE=1

  # If you use don't use certificate based authentication, but rely on username and password
  export ICINGA2_API_USERNAME=""
  export ICINGA2_API_PASSWORD=""

  # Lowest state a service has to be in, in order to be shown on the dashboad.
  # Possible values
  # 0 => OK
  # 1 => Warning
  # 2 => Critical
  # 3 => Unknown
  # This value can be overwritten by the query parameter "minState=" when opening the dashboard in a browser.
  export MIN_STATE=1

  # Highest state a service/host can be in, in order to be shown on the dashboad.
  # Possible values
  # 0 => OK
  # 1 => Warning
  # 2 => Critical
  # 3 => Unknown
  # This value can be overwritten by the query parameter "maxState=" when opening the dashboard in a browser.
  export MAX_STATE=2

  # min state type a service/host should have, in order to be shown on the dashboad.
  # Possible values
  # 0 => Soft state
  # 1 => hard state
  # This value can be overwritten by the query parameter "maxStateType=" when opening the dashboard in a browser.
  export MIN_STATE_TYPE=0
```

## SwiftBar Plugin

The repository includes a SwiftBar plugin for macOS that displays Icinga alerts directly in your menu bar.

### What is it?

The SwiftBar plugin (`swiftbar/icinga.30s.sh`) queries your Icinga2 dashboard API and shows the number and details of current warnings and critical alerts as a live menu bar item. It provides direct links to your IcingaWeb2 interface for quick access and management.

### How to install SwiftBar

1. Download SwiftBar from [swiftbar.app](https://swiftbar.app/).
2. Move SwiftBar to your Applications folder and launch it.

### How to install the plugin

1. Copy the file `swiftbar/icinga.30s.sh` from this repository to your SwiftBar plugins directory (default: `~/Library/Application Support/SwiftBar/Plugins/`).
2. Make the script executable:
   ```bash
   chmod +x ~/Library/Application\ Support/SwiftBar/Plugins/icinga.30s.sh
   ```
3. Edit the script to match your Icinga2 API URL and credentials if needed (see variables at the top of the script).
4. Ensure you have `jq` and `curl` installed (the script will auto-detect `jq`).

### Usage

- The plugin will automatically refresh every 30 seconds (as indicated by the filename `icinga.30s.sh`).
- Click on the menu bar item to see a list of current problems and direct links to acknowledge or schedule downtime in IcingaWeb2.
- If your VPN is not connected or the API is unreachable, the plugin will indicate this in the menu bar.

### Troubleshooting

- Make sure your API endpoint is reachable from your Mac.
- The script requires `jq` and `curl` to be installed and available in your PATH.
- For more details, see comments in the script file itself.
