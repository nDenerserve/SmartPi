// Package update implements SmartPi's self-update and package-management
// backend: installing an uploaded smartpi .deb, and searching/installing/
// upgrading packages from the repositories already configured on the device
// (see the readme.md installation steps - influxdata, grafana, the Raspbian
// base repos, ...).
//
// Every privileged step (apt-get, systemd-run) runs via sudo, the same
// pattern linuxtoolsRepository.ChangePassword already relies on for
// chpasswd - the smartpiserver process itself runs unprivileged as the
// "smartpi" user (see etc/systemd/system/smartpiserver.service), so a
// sudoers rule (etc/sudoers.d/smartpi-update) grants it exactly these
// commands, nothing else.
package update

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// commandInC builds a command that runs with LC_ALL=C - and thus LANGUAGE
// unset, which would otherwise override LC_ALL for gettext lookups - so its
// output is always in the untranslated, English form every parser in this
// file expects, regardless of the system's configured locale. Without this,
// a device set up with e.g. German as its locale would have apt translate
// strings this package matches literally, such as "[upgradable from: ...]"
// in Upgradable, silently turning every match into no match at all.
func commandInC(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	env := make([]string, 0, len(os.Environ())+1)
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "LC_ALL=") || strings.HasPrefix(kv, "LANGUAGE=") {
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = append(env, "LC_ALL=C")
	return cmd
}

// DefaultStagingDir is where uploaded .deb files are held until apt-get has
// consumed them, unless overridden by config.SmartPiConfig.UpdateStagingDir.
// It deliberately does not live under /var/tmp: the readme.md setup mounts
// that as a tmpfs sized for logs and small scratch files (20-30M by default,
// see its "Create tmpfs in /etc/fstab" section), which a real SmartPi
// release package - several statically linked Go binaries in one .deb -
// does not reliably fit in. /var/smartpi is the same persistent, non-tmpfs
// storage devicetoken.DefaultPath already uses for tokens.json.
const DefaultStagingDir = "/var/smartpi/update-uploads"

// logDir and stateFile also live on /var/smartpi: both need to survive the
// very service restart an update may trigger, so that the status endpoint
// can still report how the last job ended once smartpiserver comes back up.
const (
	logDir    = "/var/smartpi/update-logs"
	stateFile = "/var/smartpi/update-job.json"
)

// packageNameRE matches a syntactically valid Debian package name (Debian
// Policy §5.6.7). It is used to sanity-check package names before they are
// ever placed on an apt-get command line, even though exec.Command already
// passes each argument through untouched by any shell.
var packageNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]*$`)

// ValidPackageName reports whether name is a syntactically valid Debian
// package name.
func ValidPackageName(name string) bool {
	return packageNameRE.MatchString(name)
}

// IsAllowedDebPackage reports whether name may be installed through the .deb
// upload endpoint. Uploads are deliberately restricted to SmartPi's own
// packages - "smartpi" itself, or a "smartpi-"-prefixed split package such as
// a future "smartpi-modules" - so that the upload form can never be used to
// sideload arbitrary software. Installing anything else still remains
// possible through the apt endpoints below, which only ever reach packages
// that are already listed in a repository configured on the device.
func IsAllowedDebPackage(name string) bool {
	return name == "smartpi" || strings.HasPrefix(name, "smartpi-")
}

// InspectDeb reads the package name and version out of a .deb file's control
// data, without installing it.
func InspectDeb(path string) (name, version string, err error) {
	out, err := commandInC("dpkg-deb", "-f", path, "Package", "Version").Output()
	if err != nil {
		return "", "", fmt.Errorf("reading package metadata from %s: %w", filepath.Base(path), err)
	}

	fields := parseControlFields(string(out))
	name, version = fields["Package"], fields["Version"]
	if name == "" || version == "" {
		return "", "", fmt.Errorf("could not determine package name and version from %s", filepath.Base(path))
	}
	return name, version, nil
}

// parseControlFields parses the "Field: value" lines dpkg-deb -f prints -
// unlike, say, dpkg-query's -f format strings, it always echoes the field
// name too, not just the value - into a name to value map.
func parseControlFields(out string) map[string]string {
	fields := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), ":")
		if !found {
			continue
		}
		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return fields
}

// CleanStaleUploads removes every .deb file left behind in dir (the
// configured staging directory) by a previous upload, as long as no install
// job is currently running. It is best-effort (errors are simply ignored)
// and meant to be called before staging a new upload: now that staging
// happens on persistent storage rather than tmpfs (see DefaultStagingDir),
// an abandoned upload would otherwise sit there forever instead of being
// reclaimed by the next reboot.
func CleanStaleUploads(dir string) {
	if status, err := CurrentStatus(); err != nil || status.State == "running" {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".deb") {
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// InstalledVersion returns the currently installed version of pkg, or "" if
// it is not installed. A lookup failure - the normal case for a package that
// was never installed - is deliberately not an error: dpkg-query exits
// non-zero and writes "no packages found" to stderr for it, which is exactly
// as informative as an empty result to every caller here.
func InstalledVersion(pkg string) string {
	out, err := commandInC("dpkg-query", "-W", "-f=${Version}", pkg).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Job describes one install run, persisted across process restarts (see
// logDir/stateFile above) so its outcome can still be reported once
// smartpiserver comes back up.
type Job struct {
	// Unit is the transient systemd unit the install runs under. Each job
	// gets a fresh, unique unit name rather than reusing one, since a
	// systemd unit stays loaded (and its name unavailable for reuse) after
	// it exits until something explicitly resets or garbage-collects it -
	// see StartInstall.
	Unit string `json:"unit"`
	// Kind is "deb" for an uploaded package, "apt" for one installed or
	// upgraded from a configured repository, or "apt-all" for upgrading
	// every upgradable package at once (see StartUpgradeAll).
	Kind            string `json:"kind"`
	Package         string `json:"package,omitempty"`
	PreviousVersion string `json:"previousVersion,omitempty"`
	TargetVersion   string `json:"targetVersion,omitempty"`
	// Packages lists every package being upgraded, for Kind "apt-all" -
	// Package/PreviousVersion/TargetVersion don't apply when many packages
	// are upgraded in one shot. Best-effort: left empty if the package list
	// could not be determined up front, which does not stop the upgrade
	// itself from running.
	Packages  []string  `json:"packages,omitempty"`
	StartedAt time.Time `json:"startedAt"`
}

// Status is a Job plus its current outcome, as reported by systemd and the
// job's captured log.
type Status struct {
	Job
	// State is "idle" (no job has ever run), "running", "succeeded" or
	// "failed".
	State string `json:"state"`
	// ExitCode is apt-get's exit status, meaningful only once State is
	// "succeeded" or "failed".
	ExitCode int `json:"exitCode"`
	// Log is the tail of the job's combined stdout/stderr, populated only
	// while State is "running" - once a job has finished, its log is no
	// longer needed and would otherwise linger in every future status
	// response as confusingly stale output from a job that is long over.
	Log string `json:"log,omitempty"`
}

// jobMu serializes startJob against itself: two concurrent requests must
// not both pass the "nothing is running" check and race to launch two
// jobs at once.
var jobMu sync.Mutex

// StartInstall launches `apt-get install -y <target>` as its own detached
// systemd unit (see startJob) and returns immediately; poll CurrentStatus
// for the outcome.
//
// target is either a filesystem path to a .deb file, or an apt package
// reference ("name" or "name=version"). pkg/previousVersion/targetVersion
// are recorded purely for display - they play no part in what gets
// installed.
func StartInstall(kind, pkg, previousVersion, targetVersion, target string) (Job, error) {
	return startJob(Job{
		Kind:            kind,
		Package:         pkg,
		PreviousVersion: previousVersion,
		TargetVersion:   targetVersion,
	}, []string{"install", "-y", "-o", "Dpkg::Options::=--force-confold", target})
}

// StartUpgradeAll launches `apt-get upgrade -y`, upgrading every package
// that currently has a newer version available in the repositories
// configured on the device (as of the last Refresh) - the same set
// Upgradable reports. Like StartInstall, it returns immediately; poll
// CurrentStatus for the outcome.
func StartUpgradeAll() (Job, error) {
	job := Job{Kind: "apt-all"}
	if pkgs, err := Upgradable(); err == nil {
		names := make([]string, len(pkgs))
		for i, p := range pkgs {
			names[i] = p.Name
		}
		job.Packages = names
	}
	return startJob(job, []string{"upgrade", "-y", "-o", "Dpkg::Options::=--force-confold"})
}

// varTmpPath is where apt-get/dpkg keep scratch files - archive extraction,
// etc. - while installing or upgrading packages.
const varTmpPath = "/var/tmp"

// varTmpEnlargedSize is the size aptGetScript remounts varTmpPath to, on a
// system where it turns out to be a tmpfs, before running apt-get.
const varTmpEnlargedSize = "200m"

// aptGetScript builds the shell command startJob runs as its systemd unit's
// main process: `apt-get <aptArgs...>`, wrapped so that if varTmpPath is
// currently mounted as a tmpfs - the readme.md setup instructions have
// installers do this, sized at only 20-30M by default (see
// DefaultStagingDir's doc comment) - it is temporarily remounted large
// enough for apt-get to download and unpack many packages at once (as
// StartUpgradeAll can do), and remounted back to its normal, fstab-
// configured size again once apt-get exits. A system that leaves varTmpPath
// on disk is unaffected: the tmpfs check is a no-op there and nothing is
// remounted.
//
// This lives inside the unit itself, rather than bracketing the
// systemd-run call below in Go, because the update this triggers can
// restart smartpiserver.service mid-job - installing "smartpi" itself does
// exactly that - and the shrink-back step needs to run regardless of
// whether the Go process that started the job is still around to see it
// finish.
func aptGetScript(aptArgs []string) string {
	quoted := make([]string, len(aptArgs))
	for i, a := range aptArgs {
		quoted[i] = shellQuote(a)
	}
	aptGetCmd := "apt-get " + strings.Join(quoted, " ")

	return fmt.Sprintf(`fstype=$(awk '$2 == "%[1]s" {print $3}' /proc/mounts)
if [ "$fstype" = tmpfs ]; then
  mount -o remount,size=%[2]s %[1]s
  %[3]s
  ec=$?
  mount -o remount %[1]s
  exit $ec
fi
%[3]s
`, varTmpPath, varTmpEnlargedSize, aptGetCmd)
}

// shellQuote wraps s in single quotes for safe use as one word in a POSIX
// shell command line, escaping any single quotes it contains.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// startJob is the shared implementation behind StartInstall and
// StartUpgradeAll: it launches `apt-get <aptArgs...>` (see aptGetScript) as
// its own detached systemd unit and persists job (with Unit and StartedAt
// filled in) so CurrentStatus can report its outcome.
//
// The install is deliberately not run as a plain child process of
// smartpiserver: installing "smartpi" itself can restart smartpiserver.service
// from a postinst script, and systemd's default KillMode (control-group)
// would then kill every process in that service's cgroup, apt-get included,
// corrupting the very install in progress. Running it via `systemd-run` moves
// it into its own unit/cgroup first, and `--property=KillMode=none` on that
// unit keeps it that way even if something explicitly signals it.
func startJob(job Job, aptArgs []string) (Job, error) {
	jobMu.Lock()
	defer jobMu.Unlock()

	if status, err := CurrentStatus(); err == nil && status.State == "running" {
		return Job{}, fmt.Errorf("an update is already running (unit %s)", status.Unit)
	}

	if err := os.MkdirAll(logDir, 0700); err != nil {
		return Job{}, fmt.Errorf("creating %s: %w", logDir, err)
	}

	unit := fmt.Sprintf("smartpi-update-%d", time.Now().UnixNano())
	logPath := filepath.Join(logDir, unit+".log")

	args := []string{
		"systemd-run",
		"--unit=" + unit,
		"--property=KillMode=none",
		"--property=StandardOutput=file:" + logPath,
		"--property=StandardError=file:" + logPath,
		"--",
		"/bin/sh", "-c", aptGetScript(aptArgs),
	}
	if out, err := exec.Command("sudo", args...).CombinedOutput(); err != nil {
		return Job{}, fmt.Errorf("starting update: %s (%s)", err, strings.TrimSpace(string(out)))
	}

	job.Unit = unit
	job.StartedAt = time.Now().UTC()
	if err := saveJob(job); err != nil {
		return job, fmt.Errorf("update started but its state could not be saved: %w", err)
	}
	return job, nil
}

// CurrentStatus reports the state of the most recently started job, or
// State "idle" if none has ever run on this device.
func CurrentStatus() (Status, error) {
	job, ok, err := loadJob()
	if err != nil {
		return Status{}, err
	}
	if !ok {
		return Status{State: "idle"}, nil
	}

	state, exitCode := unitState(job.Unit)

	var logTail string
	if state == "running" {
		logTail, _ = tailFile(filepath.Join(logDir, job.Unit+".log"), 64*1024)
	}

	return Status{
		Job:      job,
		State:    state,
		ExitCode: exitCode,
		Log:      logTail,
	}, nil
}

// unitStateExecRetries/unitStateExecRetryDelay bound how hard unitState
// tries to run `systemctl show` before giving up - see unitState's doc
// comment for why a single failed attempt isn't trusted.
const (
	unitStateExecRetries    = 3
	unitStateExecRetryDelay = 200 * time.Millisecond
)

// unitState queries systemd for unit's current lifecycle state. Reading a
// unit's status is an unprivileged operation, unlike starting or resetting
// one, so this deliberately does not go through sudo.
//
// Spawning systemctl itself can transiently fail - fork/exec: resource
// temporarily unavailable - while a large apt-get run (hundreds of
// packages, their postinst scripts, ...) is under way, especially on the
// memory-constrained hardware SmartPi typically runs on. A single failed
// attempt used to be reported as state "unknown" straight away, which a
// caller polling this can't tell apart from a genuinely lost job - even
// though the apt-get run itself was still very much alive. Retrying a few
// times first, rather than trusting one attempt, avoids that false
// "unknown" for what is almost always a momentary blip.
func unitState(unit string) (state string, exitCode int) {
	var out []byte
	var err error
	for attempt := 0; attempt < unitStateExecRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(unitStateExecRetryDelay)
		}
		out, err = exec.Command("systemctl", "show", unit,
			"--property=ActiveState", "--property=SubState",
			"--property=Result", "--property=ExecMainStatus",
		).Output()
		if err == nil {
			break
		}
	}
	if err != nil {
		return "unknown", 0
	}

	props := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		k, v, found := strings.Cut(scanner.Text(), "=")
		if found {
			props[k] = v
		}
	}

	fmt.Sscanf(props["ExecMainStatus"], "%d", &exitCode)

	switch props["ActiveState"] {
	case "activating", "reloading":
		return "running", exitCode
	case "active":
		if props["SubState"] == "running" {
			return "running", exitCode
		}
	}

	if props["Result"] == "success" {
		return "succeeded", exitCode
	}
	if props["ActiveState"] == "" {
		// The unit is no longer loaded at all (e.g. the device rebooted
		// mid-install, or something ran `systemctl reset-failed`). Neither
		// "succeeded" nor "failed" would be accurate, since we genuinely
		// don't know.
		return "unknown", exitCode
	}
	return "failed", exitCode
}

// saveJob persists job atomically: to a temp file in stateFile's directory,
// then renamed into place, mirroring the pattern devicetoken.Store.save uses
// for tokens.json.
func saveJob(job Job) error {
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding job state: %w", err)
	}

	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".update-job-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpPath, stateFile); err != nil {
		return fmt.Errorf("replacing %s: %w", stateFile, err)
	}
	return nil
}

// loadJob reads back the last job saved by saveJob. ok is false if no job
// has ever run on this device.
func loadJob() (job Job, ok bool, err error) {
	data, err := os.ReadFile(stateFile)
	if os.IsNotExist(err) {
		return Job{}, false, nil
	}
	if err != nil {
		return Job{}, false, fmt.Errorf("reading %s: %w", stateFile, err)
	}
	if err := json.Unmarshal(data, &job); err != nil {
		return Job{}, false, fmt.Errorf("parsing %s: %w", stateFile, err)
	}
	return job, true, nil
}

// tailFile returns the last maxBytes of the file at path, or its entirety if
// smaller. Missing file (no output has been captured yet) is not an error.
func tailFile(path string, maxBytes int64) (string, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", err
	}

	size := info.Size()
	offset := int64(0)
	truncated := false
	if size > maxBytes {
		offset = size - maxBytes
		truncated = true
	}
	if _, err := f.Seek(offset, 0); err != nil {
		return "", err
	}

	buf := make([]byte, size-offset)
	if _, err := io.ReadFull(f, buf); err != nil {
		return "", err
	}

	if truncated {
		return "... (truncated)\n" + string(buf), nil
	}
	return string(buf), nil
}

// Refresh runs `apt-get update`, refreshing the package index from every
// repository already configured on the device. It runs synchronously - it
// never restarts a service, so it carries none of StartInstall's risk - and
// returns its combined output for display.
func Refresh() (string, error) {
	out, err := commandInC("sudo", "apt-get", "update").CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("apt-get update failed: %s (%s)", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// PackageSummary is one hit from Search.
type PackageSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Search looks up term against every package name known from the configured
// repositories (apt-cache search, restricted to names so a package's
// description text can't produce surprising matches).
func Search(term string) ([]PackageSummary, error) {
	out, err := commandInC("apt-cache", "search", "--names-only", term).Output()
	if err != nil {
		return nil, fmt.Errorf("apt-cache search failed: %w", err)
	}
	return parseSearchOutput(string(out)), nil
}

func parseSearchOutput(out string) []PackageSummary {
	// Initialized rather than nil so a genuinely empty result still encodes
	// to "[]", not "null" - the same reasoning as parseUpgradableOutput.
	results := []PackageSummary{}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		name, desc, _ := strings.Cut(line, " - ")
		results = append(results, PackageSummary{Name: name, Description: desc})
	}
	return results
}

// PackageInfo is one package's installed vs. available version, as reported
// by the repositories configured on the device.
type PackageInfo struct {
	Name             string `json:"name"`
	InstalledVersion string `json:"installedVersion,omitempty"`
	CandidateVersion string `json:"candidateVersion,omitempty"`
}

// Info reports name's installed and candidate (best available) version.
func Info(name string) (PackageInfo, error) {
	out, err := commandInC("apt-cache", "policy", name).Output()
	if err != nil {
		return PackageInfo{}, fmt.Errorf("apt-cache policy failed: %w", err)
	}
	return parsePolicyOutput(name, string(out)), nil
}

func parsePolicyOutput(name, out string) PackageInfo {
	info := PackageInfo{Name: name}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "Installed:"):
			if v := strings.TrimSpace(strings.TrimPrefix(line, "Installed:")); v != "(none)" {
				info.InstalledVersion = v
			}
		case strings.HasPrefix(line, "Candidate:"):
			if v := strings.TrimSpace(strings.TrimPrefix(line, "Candidate:")); v != "(none)" {
				info.CandidateVersion = v
			}
		}
	}
	return info
}

// UpgradablePackage is one package with a pending upgrade.
type UpgradablePackage struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"currentVersion"`
	NewVersion     string `json:"newVersion"`
}

// upgradableLineRE matches one data line of `apt list --upgradable`, e.g.:
// "smartpi/stable 1.2.4 armhf [upgradable from: 1.2.3]"
var upgradableLineRE = regexp.MustCompile(`^(\S+)/\S+\s+(\S+)\s+\S+\s+\[upgradable from:\s*([^\]]+)\]`)

// Upgradable lists every package with a newer version available in the
// repositories configured on the device (as of the last Refresh).
func Upgradable() ([]UpgradablePackage, error) {
	// apt (unlike apt-get/apt-cache) prints a stability warning to stderr on
	// every invocation ("apt does not have a stable CLI interface") - not a
	// real error, so it is deliberately discarded here rather than folded
	// into err via CombinedOutput.
	out, err := commandInC("apt", "list", "--upgradable").Output()
	if err != nil {
		return nil, fmt.Errorf("apt list --upgradable failed: %w", err)
	}
	return parseUpgradableOutput(string(out)), nil
}

func parseUpgradableOutput(out string) []UpgradablePackage {
	// Initialized rather than nil so a genuinely empty result still encodes
	// to "[]", not "null" - which reads exactly like the locale bug this
	// package used to have (see commandInC), where every line silently
	// failed to match and looked the same as "nothing to upgrade".
	results := []UpgradablePackage{}
	for _, line := range strings.Split(out, "\n") {
		m := upgradableLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		results = append(results, UpgradablePackage{
			Name:           m[1],
			NewVersion:     m[2],
			CurrentVersion: m[3],
		})
	}
	return results
}
