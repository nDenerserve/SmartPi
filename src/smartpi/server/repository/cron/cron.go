/*
    Copyright (C) Jens Ramhorst
	  This file is part of SmartPi.
    SmartPi is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    SmartPi is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.
    Diese Datei ist Teil von SmartPi.
    SmartPi ist Freie Software: Sie können es unter den Bedingungen
    der GNU General Public License, wie von der Free Software Foundation,
    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
    veröffentlichten Version, weiterverbreiten und/oder modifizieren.
    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
    Siehe die GNU General Public License für weitere Details.
    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
*/

// Package cronRepository keeps /etc/cron.d/smartpi's smartpiftpupload line in
// sync with the FTP settings in /etc/smartpi (see [ftp] ftp_upload and
// ftp_sendtimes in smartpi/config/smartpiconfig_file.go): commented out while
// FTP upload is disabled, uncommented with a minute/hour schedule matching
// the configured send times while it is enabled. See CronFilePath's doc
// comment for the line format this rewrites.
package cronRepository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nDenerserve/SmartPi/smartpi/rootcall"
)

// CronFilePath is the cron.d file smartpiftpupload's schedule lives in. A
// fresh copy - commented out - ships at etc/cron.d/smartpi in this
// repository and is installed the same manual way as the systemd units next
// to it (see readme.md); SyncFTPUpload below also creates it from scratch if
// it is ever missing, so a first boot without that manual step still ends up
// with a working, if initially disabled, file.
const CronFilePath = "/etc/cron.d/smartpi"

// ftpUploadUser/ftpUploadCommand identify smartpiftpupload's line among any
// others a device operator may have added to the same file by hand -
// SyncFTPUpload only ever touches the one line containing ftpUploadCommand,
// every other line (comments, blank lines, unrelated jobs) is left exactly
// as it was.
const ftpUploadUser = "smartpi"
const ftpUploadCommand = "/usr/local/bin/smartpiftpupload"

type CronRepository struct{}

// SyncFTPUpload rewrites CronFilePath's smartpiftpupload line to match
// enabled/sendtimes: commented out (a no-op line, same as the file ships)
// when disabled, or an active "0 <hours> * * *" schedule when enabled - see
// buildFTPUploadLine. It is called with the same config values every time
// /etc/smartpi is written or (re)loaded (see the config controller and
// server.go's startup path), so it is a plain last-write-wins sync, not an
// incremental edit.
func (c CronRepository) SyncFTPUpload(enabled bool, sendtimes [24]bool) error {
	state := "off"
	if enabled {
		state = "on"
	}
	// /etc/cron.d/smartpi is root-owned: the root helper replaces the line
	// (see WriteFTPUpload), smartpiserver itself runs unprivileged
	out, err := rootcall.Command(rootcall.CronFTP, state, hourField(sendtimes)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func buildFTPUploadLine(enabled bool, sendtimes [24]bool) string {
	return ftpLine(enabled, hourField(sendtimes))
}

func ftpLine(enabled bool, hours string) string {
	prefix := ""
	if !enabled {
		prefix = "#"
	}
	return fmt.Sprintf("%s0 %s * * * %s  %s", prefix, hours, ftpUploadUser, ftpUploadCommand)
}

// hourRE matches what hourField produces: "*" or a comma list of hours.
var hourRE = regexp.MustCompile(`^(\*|([0-9]|1[0-9]|2[0-3])(,([0-9]|1[0-9]|2[0-3])){0,23})$`)

// ValidHours reports whether hours is a cron hour field hourField produces.
func ValidHours(hours string) bool {
	return hourRE.MatchString(hours)
}

// WriteFTPUpload replaces the smartpiftpupload line of CronFilePath (or
// adds it) and leaves every other line alone. It runs as root in the root
// helper; the hour field is checked here because the caller may be anyone
// acting as the smartpi user, and the line always runs the fixed command as
// the fixed user.
func WriteFTPUpload(enabled bool, hours string) error {
	return writeFTPUpload(CronFilePath, enabled, hours)
}

func writeFTPUpload(path string, enabled bool, hours string) error {
	if !ValidHours(hours) {
		return fmt.Errorf("invalid hours %q", hours)
	}
	lines, err := readLines(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	newLine := ftpLine(enabled, hours)
	replaced := false
	for i, line := range lines {
		if strings.Contains(line, ftpUploadCommand) {
			lines[i] = newLine
			replaced = true
		}
	}
	if !replaced {
		lines = append(lines, newLine)
	}
	// written next to the target and renamed: cron ignores file names with
	// a dot, so it never sees a half written file
	tmp := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".tmp")
	if err := os.WriteFile(tmp, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0644); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// hourField turns the 24 hourly on/off flags from the settings "FTP" tab
// into a cron hour field: "*" (every hour) when all 24 are set - the
// "hourly" quick choice in the UI - a comma list of the selected hours for
// anything in between, and "0" (midnight only, the file's original default)
// if somehow none are set, so a fully-cleared selection still produces a
// valid schedule rather than an invalid empty field.
func hourField(sendtimes [24]bool) string {
	hours := make([]string, 0, 24)
	for h := 0; h < 24; h++ {
		if sendtimes[h] {
			hours = append(hours, strconv.Itoa(h))
		}
	}
	if len(hours) == 0 {
		return "0"
	}
	if len(hours) == 24 {
		return "*"
	}
	return strings.Join(hours, ",")
}

func readLines(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := strings.TrimRight(string(b), "\n")
	if content == "" {
		return []string{}, nil
	}
	return strings.Split(content, "\n"), nil
}
