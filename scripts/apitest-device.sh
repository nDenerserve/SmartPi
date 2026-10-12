#!/bin/bash
# Functional test of a SmartPi's web API after an update.
#
#   scripts/apitest-device.sh [host] [--upload file.deb]
#
# The password of the user smartpi comes from SMARTPI_PASS or, if that is
# not set, from ~/.config/smartpi-test/password (mode 600).
#
# Creates the user "sptest" (removed again at the end), saves the settings
# once with an FTP hour toggled and once restored, and with --upload
# installs the given package through the Update tab's upload.
# Needs curl, jq, ssh and sshpass (ssh login as smartpi with SMARTPI_PASS).
set -u
HOST=${1:-10.1.0.171}
UPLOAD=""
[ "${2:-}" = "--upload" ] && UPLOAD=${3:?package file missing}
PASSFILE=~/.config/smartpi-test/password
[ -z "${SMARTPI_PASS:-}" ] && [ -r "$PASSFILE" ] && SMARTPI_PASS=$(head -n1 "$PASSFILE")
: "${SMARTPI_PASS:?set SMARTPI_PASS or create $PASSFILE}"
API=http://$HOST:1080/api/v1
export SSHPASS=$SMARTPI_PASS
FAILED=0

ok()   { echo "  OK    $*"; }
fail() { echo "  FAIL  $*"; FAILED=1; }
check() { # check <description> <expected> <actual>
	if [ "$2" = "$3" ]; then ok "$1 ($3)"; else fail "$1: expected $2, got $3"; fi
}
dev() { sshpass -e ssh -o StrictHostKeyChecking=accept-new "smartpi@$HOST" "$@"; }
status() { curl -s -o /dev/null -w '%{http_code}' "$@"; }
login() { # login <user> <password> -> prints the session token
	jq -n --arg u "$1" --arg p "$2" '{username:$u,password:$p}' |
		curl -s -X POST "$API/login" -H 'Content-Type: application/json' -d @- | jq -r '.token // empty'
}
login_status() {
	jq -n --arg u "$1" --arg p "$2" '{username:$u,password:$p}' |
		curl -s -o /dev/null -w '%{http_code}' -X POST "$API/login" -H 'Content-Type: application/json' -d @-
}
# remote sudo reads the password from stdin, never from the command line
rsudo() { printf '%s\n' "$SMARTPI_PASS" | dev "sudo -S -p '' $*"; }

echo "== Login"
TOKEN=$(login smartpi "$SMARTPI_PASS")
[ -n "$TOKEN" ] && ok "login smartpi" || { fail "login smartpi"; exit 1; }
AUTH=(-H "Authorization: Bearer $TOKEN")
check "wrong password" 401 "$(login_status smartpi wrong-password)"
check "invalid user name" 401 "$(login_status 'smart pi' x)"
check "users without token" 401 "$(status "$API/users")"
check "users with token" 200 "$(status "${AUTH[@]}" "$API/users")"

echo "== User management"
TPW1="T1-$(openssl rand -hex 8)"
TPW2="T2-$(openssl rand -hex 8)"
check "create sptest" 201 "$(jq -n --arg p "$TPW1" '{username:"sptest",password:$p}' |
	curl -s -o /dev/null -w '%{http_code}' -X POST "${AUTH[@]}" "$API/users" -H 'Content-Type: application/json' -d @-)"
curl -s "${AUTH[@]}" "$API/users" | jq -e '[.[].username] | index("sptest")' >/dev/null &&
	ok "sptest listed" || fail "sptest not listed"
check "login sptest" 200 "$(login_status sptest "$TPW1")"
check "change sptest password" 200 "$(jq -n --arg p "$TPW2" '{password:$p}' |
	curl -s -o /dev/null -w '%{http_code}' -X POST "${AUTH[@]}" "$API/users/sptest/password" -H 'Content-Type: application/json' -d @-)"
check "login sptest new password" 200 "$(login_status sptest "$TPW2")"
check "login sptest old password" 401 "$(login_status sptest "$TPW1")"
check "invalid user name rejected" 400 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "${AUTH[@]}" "$API/users" \
	-H 'Content-Type: application/json' -d '{"username":"a;b","password":"irrelevant-123"}')"
rsudo userdel -r sptest 2>/dev/null && ok "sptest removed" || fail "removing sptest"

echo "== Settings (password fields, FTP schedule)"
CONF=$(curl -s "${AUTH[@]}" "$API/config/readsmartpiconfiguration")
for f in AppKey FTPpass MQTTpass SmartpicloudMQTTpass Influxpassword InfluxAPIToken; do
	v=$(jq -r --arg f "$f" '.[$f] // empty' <<<"$CONF")
	[ -z "$v" ] || [ "$v" = "********" ] && ok "$f masked" || fail "$f is returned in plain text"
done
secrets_hash() { rsudo grep -E "'^(appkey|ftp_pass|mqtt_password|smartpicloud_mqtt_password|influxpassword|influxapitoken)'" /etc/smartpi | sha256sum; }
BEFORE=$(secrets_hash)
# the settings view posts {"type":"config","msg":<configuration>}
NEWCONF=$(jq '{type:"config", msg:(.FTPupload = true | .FTPsendtimes[3] = (.FTPsendtimes[3] | not))}' <<<"$CONF")
OLDCONF=$(jq '{type:"config", msg:.}' <<<"$CONF")
check "save settings (hour 3 toggled)" 200 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "${AUTH[@]}" \
	"$API/config/writesmartpiconfiguration" -H 'Content-Type: application/json' -d "$NEWCONF")"
check "masked secrets kept" "$BEFORE" "$(secrets_hash)"
echo "  cron.d/smartpi after saving:"; dev cat /etc/cron.d/smartpi | grep -v '^#' | sed 's/^/        /'
check "config mode" 640 "$(dev stat -c %a /etc/smartpi)"
check "restore settings" 200 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "${AUTH[@]}" \
	"$API/config/writesmartpiconfiguration" -H 'Content-Type: application/json' -d "$OLDCONF")"
check "secrets after restore" "$BEFORE" "$(secrets_hash)"
echo "  cron.d/smartpi restored:"; dev cat /etc/cron.d/smartpi | grep -v '^#' | sed 's/^/        /'

echo "== Update tab"
check "version" 200 "$(status "${AUTH[@]}" "$API/update/version")"
curl -s "${AUTH[@]}" "$API/update/version" | sed 's/^/        /'; echo
check "apt refresh" 200 "$(status -X POST "${AUTH[@]}" "$API/apt/refresh")"
check "apt search" 200 "$(status "${AUTH[@]}" "$API/apt/search?q=smartpi")"
check "apt package info" 200 "$(status "${AUTH[@]}" "$API/apt/package/smartpi")"
check "apt upgradable" 200 "$(status "${AUTH[@]}" "$API/apt/upgradable")"
check "update status" 200 "$(status "${AUTH[@]}" "$API/update/status")"
if [ -n "$UPLOAD" ]; then
	check "upload $(basename "$UPLOAD")" 202 "$(status -X POST "${AUTH[@]}" -F "file=@$UPLOAD" "$API/update/package")"
	for i in $(seq 1 60); do
		sleep 5
		S=$(curl -s --max-time 5 "${AUTH[@]}" "$API/update/status" 2>/dev/null) || continue
		jq -e '.state == "running"' >/dev/null 2>&1 <<<"$S" || break
	done
	echo "  update status: $S"
	dev 'for s in smartpireadout smartpiserver smartpiemeter smartpimodbus; do echo "        $s: $(systemctl show -p SubState --value $s)"; done'
fi

echo "== Charts and CSV"
check "progressdata" 200 "$(status "$API/smartpiac/progressdata/value/power/aggregate/1h")"
check "barchart" 200 "$(status "$API/smartpiac/barchart/value/energy_pos/aggregate/1d")"
S=$(date -u -d '-2 hours' +%Y-%m-%dT%H:%M:%S.000Z); E=$(date -u +%Y-%m-%dT%H:%M:%S.000Z)
check "csv export" 200 "$(status "$API/smartpiac/csvexport/start/$S/stop/$E/aggregate/15m")"
check "invalid aggregate" 400 "$(status "$API/smartpiac/csvexport/start/$S/stop/$E/aggregate/x)")"

echo; [ $FAILED = 0 ] && echo "ALL TESTS PASSED" || echo "SOME TESTS FAILED"
exit $FAILED
