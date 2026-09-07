package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	"github.com/nDenerserve/SmartPi/smartpi/update"
)

// maxUploadBytes bounds the size of an uploaded .deb, generously - while
// still keeping a runaway or malicious upload from filling the staging
// directory's filesystem (config.SmartPiConfig.UpdateStagingDir).
const maxUploadBytes = 200 << 20

// updatePackageName is the package SmartPi's own release .deb is expected to
// use. Version() reports its installed version alongside the running
// binary's own build version.
const updatePackageName = "smartpi"

// GetUpdateVersion reports the version of the currently running smartpiserver
// binary (baked in at build time, see the makefile's -ldflags) alongside the
// version dpkg currently has on record for the smartpi package. The two can
// disagree - most commonly right after an update whose install has not
// finished yet, or on a device whose binaries were replaced by hand rather
// than through a package - which is exactly why both are reported rather
// than just one.
func (c Controller) GetUpdateVersion(appVersion string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverutils.ResponseJSON(w, struct {
			RunningVersion   string `json:"runningVersion"`
			InstalledVersion string `json:"installedVersion,omitempty"`
		}{
			RunningVersion:   appVersion,
			InstalledVersion: update.InstalledVersion(updatePackageName),
		})
	}
}

// UploadUpdatePackage handles a .deb upload: it inspects the file, rejects
// anything that is not a SmartPi package (see update.IsAllowedDebPackage),
// and starts installing it. The HTTP response only confirms the install was
// started - poll GetUpdateStatus for its outcome, since installing "smartpi"
// itself can restart smartpiserver mid-request.
func (c Controller) UploadUpdatePackage(conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		stagingDir := conf.UpdateStagingDir

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			errorObject.Message = "Uploaded file is too large or malformed."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			errorObject.Message = "Missing uploaded file (expected multipart field \"file\")."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}
		defer file.Close()

		if !strings.HasSuffix(strings.ToLower(header.Filename), ".deb") {
			errorObject.Message = "Only .deb package files are supported."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		if err := os.MkdirAll(stagingDir, 0700); err != nil {
			log.Errorf("update: creating staging directory %s: %v", stagingDir, err)
			errorObject.Message = "Could not prepare the upload directory."
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}
		update.CleanStaleUploads(stagingDir)

		// Each upload gets its own filename rather than a fixed one, so a
		// second upload arriving while a previous install is still starting
		// up can never overwrite the file that install is reading.
		destPath := filepath.Join(stagingDir, fmt.Sprintf("upload-%d.deb", time.Now().UnixNano()))
		if err := saveUploadedFile(file, destPath); err != nil {
			log.Errorf("update: storing upload at %s: %v", destPath, err)
			errorObject.Message = fmt.Sprintf("Could not store the uploaded file: %v", err)
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}

		pkgName, version, err := update.InspectDeb(destPath)
		if err != nil {
			os.Remove(destPath)
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		if !update.IsAllowedDebPackage(pkgName) {
			os.Remove(destPath)
			errorObject.Message = fmt.Sprintf("Package %q cannot be installed through upload; only smartpi packages are accepted.", pkgName)
			serverutils.RespondWithError(w, http.StatusForbidden, errorObject)
			return
		}

		previousVersion := update.InstalledVersion(pkgName)

		job, err := update.StartInstall("deb", pkgName, previousVersion, version, destPath)
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusConflict, errorObject)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		serverutils.ResponseJSON(w, job)
	}
}

func saveUploadedFile(src io.Reader, destPath string) error {
	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dest, src); err != nil {
		dest.Close()
		os.Remove(destPath)
		return err
	}
	return dest.Close()
}

// GetUpdateStatus reports the state of the most recently started install
// job, whether it was triggered by UploadUpdatePackage or InstallAptPackage.
func (c Controller) GetUpdateStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		status, err := update.CurrentStatus()
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}
		serverutils.ResponseJSON(w, status)
	}
}

// RefreshAptCache runs apt-get update against the repositories already
// configured on the device, so Search/Info/ListUpgradablePackages reflect
// current versions.
func (c Controller) RefreshAptCache() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		out, err := update.Refresh()
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}
		serverutils.ResponseJSON(w, struct {
			Output string `json:"output"`
		}{Output: out})
	}
}

// SearchAptPackages looks up ?q= against every package name known from the
// repositories configured on the device.
func (c Controller) SearchAptPackages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		term := r.URL.Query().Get("q")
		if term == "" {
			errorObject.Message = "Missing query parameter \"q\"."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		results, err := update.Search(term)
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}
		serverutils.ResponseJSON(w, results)
	}
}

// GetAptPackageInfo reports one package's installed and candidate version.
func (c Controller) GetAptPackageInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		name := mux.Vars(r)["name"]
		if !update.ValidPackageName(name) {
			errorObject.Message = "Invalid package name."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		info, err := update.Info(name)
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}
		serverutils.ResponseJSON(w, info)
	}
}

// ListUpgradablePackages lists every package with a newer version available,
// as of the last RefreshAptCache.
func (c Controller) ListUpgradablePackages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		results, err := update.Upgradable()
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}
		serverutils.ResponseJSON(w, results)
	}
}

// installAptPackageRequest is the body of POST /api/v1/apt/install.
type installAptPackageRequest struct {
	Package string `json:"package"`
	// Version pins a specific candidate; left empty, apt-get installs
	// whatever the configured repositories currently offer as newest.
	Version string `json:"version,omitempty"`
}

// InstallAptPackage installs (or upgrades) one package from the
// repositories already configured on the device. Like UploadUpdatePackage,
// this only starts the job - poll GetUpdateStatus for its outcome.
func (c Controller) InstallAptPackage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		var req installAptPackageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorObject.Message = "Malformed request body."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}
		if !update.ValidPackageName(req.Package) {
			errorObject.Message = "Invalid package name."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		target := req.Package
		targetVersion := req.Version
		if req.Version != "" {
			target = req.Package + "=" + req.Version
		} else if info, err := update.Info(req.Package); err == nil {
			targetVersion = info.CandidateVersion
		}

		previousVersion := update.InstalledVersion(req.Package)

		job, err := update.StartInstall("apt", req.Package, previousVersion, targetVersion, target)
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusConflict, errorObject)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		serverutils.ResponseJSON(w, job)
	}
}
