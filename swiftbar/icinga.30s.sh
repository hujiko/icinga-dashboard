#!/bin/bash
set -euo pipefail

# ------------------------------------------------------------------------------
# icinga.30s.sh – SwiftBar plugin for Icinga2 monitoring
#
# Shows critical and warning services/hosts from an Icinga2 API in the menu bar.
#
# Requirements:
#   - Bash
#   - curl
#   - jq (auto-detected)
#   - Network access to the Icinga-Dashboard API
#
# Configuration:
#   - Adjust ICINGAWEBURL, APIURL, etc. in the script
#
# Output:
#   - Status line with count of critical and warning services/hosts
#   - Details with links to IcingaWeb2
#
# Errors are printed to STDERR, exit code 1 on error.
# ------------------------------------------------------------------------------

# --- Constants ---
# Hostname of your icingaweb2 instance
readonly ICINGAWEBURL=icinga2.example.com

# HTTP URL of Icinga2.
readonly ICINGAWEBLINK=https://icinga2.example.com/

# Endpoint of the Icinga-Dashbaord API.
readonly APIURL="https://icinga2-dashboard.example.com/api/v1/dashboard"

# Location of the curl binary
readonly CURLPATH=/usr/bin/curl

# Colors and fonts
readonly STYLE_RED="| color=red | font=UbuntuMono-Bold"
readonly STYLE_ORANGE="| color=orange | font=UbuntuMono-Bold"

# Try to find jq automatically if not present at default path
if [ -x /opt/homebrew/bin/jq ]; then
	JQPATH=/opt/homebrew/bin/jq
elif command -v jq >/dev/null 2>&1; then
	JQPATH=$(command -v jq)
else
	echo "❌ jq not found! $STYLE_RED" >&2
	exit 1
fi

# Print an error message to STDERR and exit
# Arguments:
#   $1: Error message
error_exit() {
	local msg="$1"
	echo "❌ $msg $STYLE_RED" >&2
	exit 1
}

# Check if a command exists, else exit with error
# Arguments:
#   $1: Command name or path
check_command() {
	command -v "$1" >/dev/null 2>&1 || error_exit "$1 not found!"
}

check_command /sbin/ping
check_command "$CURLPATH"
check_command "$JQPATH"

service_show_url_base="https://$ICINGAWEBURL/icingaweb2/icingadb/service"
host_show_url_base="https://$ICINGAWEBURL/icingaweb2/icingadb/host"
downtime_url_base="https://$ICINGAWEBURL/icingaweb2/icingadb/host/schedule-downtime"
ack_url_base="https://$ICINGAWEBURL/icingaweb2/icingadb/service/acknowledge"

# --- Functions ---

# Print a service line and its acknowledge action for SwiftBar
# Arguments:
#   $1: Prefix symbol (e.g. 🛑 or 🟡)
#   $2: Service name
#   $3: Host name
print_service() {
	local prefix="$1"
	local name="$2"
	local host="$3"
	echo "$prefix $name@$host | length=100 | href=$service_show_url_base?host.name=$host&name=$name"
	echo "-- Ack $name@$host | href=$ack_url_base?host.name=$host&name=$name | alternate=true"
}

# Print a host line and its downtime action for SwiftBar
# Arguments:
#   $1: Prefix symbol (e.g. 🛑)
#   $2: Host name
print_host() {
	local prefix="$1"
	local host="$2"
	echo "$prefix Host: $host | href=$host_show_url_base?name=$host"
	echo "-- Create Downtime for $host | length=100 | href=$downtime_url_base?name=$host | alternate=true"
}

# Count services, considering aggregation (fast jq-only)
# Arguments:
#   $1: JSON string of services
# Output:
#   Echoes the total count (integer)
count_services() {
	local json="$1"
	echo "$json" | $JQPATH 'map(if .is_aggregated then (.aggregated_hosts_count // 0) else 1 end) | add // 0'
}

# Print services (both aggregated and non-aggregated) sorted alphabetically
# Arguments:
#   $1: JSON string of services
#   $2: Prefix symbol (e.g. 🛑 or 🟡)
print_services_sorted() {
	local json="$1"
	local prefix="$2"
	echo "$json" | $JQPATH -r 'sort_by(.name) | .[] | if .is_aggregated then "\(.name)\t\(.aggregated_hosts_count)\taggregated" else "\(.name)\t\(.host_field)\tnon_aggregated" end' | while IFS=$'\t' read -r name value type; do
		if [ "$type" = "aggregated" ]; then
			echo "$prefix $name ($value hosts)"
		else
			print_service "$prefix" "$name" "$value"
		fi
	done
}

# Print the status line for SwiftBar
# Arguments:
#   $1: Critical count
#   $2: Warning count
print_status_line() {
	local crit="$1"
	local warn="$2"
	if [ "$crit" -gt 0 ] && [ "$warn" -gt 0 ]; then
		echo "🛑 $crit  🟡 $warn    $STYLE_ORANGE"
	elif [ "$crit" -gt 0 ]; then
		echo "🛑 $crit    $STYLE_RED"
	elif [ "$warn" -gt 0 ]; then
		echo "🟡 $warn    $STYLE_ORANGE"
	else
		echo "✔️ No alerts"
	fi
}

# Print a separator line if needed
# Arguments:
#   $1: Boolean (true/false) whether to print separator
print_separator() {
	local show="$1"
	if [ "$show" = true ]; then
		echo "---"
	fi
}

if ! /sbin/ping -c 1 "$ICINGAWEBURL" &>/dev/null; then
	echo "🔌 NO VPN $STYLE_RED"
else
	dashboard_json=$($CURLPATH --silent --fail "$APIURL")
	if [ $? -ne 0 ] || [ -z "$dashboard_json" ]; then
		error_exit "Failed to fetch dashboard data!"
	fi
	service_warn=$(echo "$dashboard_json" | $JQPATH '.services | map(select(.state==1 and .state_type==1))')
	service_crit=$(echo "$dashboard_json" | $JQPATH '.services | map(select(.state==2 and .state_type==1))')
	host_crit=$(echo "$dashboard_json" | $JQPATH '.hosts | map(select(.state==1 and .state_type==1))')

	# Calculate warn and crit counts using functions
	local_warn=$(count_services "$service_warn")
	local_crit=$(count_services "$service_crit")
	local_host_crit=$(echo "$host_crit" | $JQPATH 'length')
	warn=$local_warn
	crit=$((local_crit + local_host_crit))

	print_status_line "$crit" "$warn"

	show_separator=false

	# Critical hosts and services
	if [ "$crit" -gt 0 ]; then
		echo "---"
		echo "🛑 $crit  $STYLE_RED"
		echo "$host_crit" | $JQPATH -r '.[] | [.name] | @tsv' | while IFS=$'\t' read -r host; do
			print_host "🛑" "$host"
		done
		print_services_sorted "$service_crit" "🛑"
		show_separator=true
	fi

	# Warnings
	if [ "$warn" -gt 0 ]; then
		if [ "$crit" -gt 0 ]; then
			echo "---"
			echo "🟡 $warn 🟡 $STYLE_ORANGE"
		fi
		print_services_sorted "$service_warn" "🟡"
		show_separator=true
	fi

	print_separator "$show_separator"
fi

echo "---"
echo "🔗 Open Icinga | href=$ICINGAWEBLINK"
echo "🔄 Refresh Now | refresh=true"
