package modulescontrollers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	modulesModels "github.com/nDenerserve/SmartPi/models/modules"
	config "github.com/nDenerserve/SmartPi/smartpi/config"
	modulesRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/modules"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// analogOut010VAllowed applies the per-user allowlist for the 0-10V output
// module. A device token already carries the analogout scope required by
// TokenVerifyMiddleWare, so - as for the 4-20mA module - the allowlist only
// applies to session tokens. Writes the error response itself and reports
// whether the handler may continue.
func analogOut010VAllowed(w http.ResponseWriter, r *http.Request, mconf *config.Moduleconfig, conf *config.SmartPiConfig) bool {
	var error models.Error
	if serverutils.IsDeviceToken(r) {
		return true
	}
	user, err := serverutils.DecryptUserdataFromToken(r, conf)
	if err != nil {
		error.Message = err.Error()
		serverutils.RespondWithError(w, http.StatusUnauthorized, error)
		return false
	}
	if !slices.Contains(mconf.AllowedAnalogOut010VUser, user.Name) {
		log.Warnf("User %s not allowed to access analogout010v (allowed: %v)", user.Name, mconf.AllowedAnalogOut010VUser)
		error.Message = "User not allowed"
		serverutils.RespondWithError(w, http.StatusUnauthorized, error)
		return false
	}
	return true
}

func respondAnalogOut010V(w http.ResponseWriter, addressStr string, status modulesModels.AnalogOut010VStatus, err error) {
	status.Moduleaddress = addressStr
	if err != nil {
		serverutils.RespondWithError(w, http.StatusInternalServerError, models.Error{Message: err.Error()})
		return
	}
	if err := json.NewEncoder(w).Encode(status); err != nil {
		panic(err)
	}
}

// parseAnalogOut010VVoltage validates a 0-10V value from a request.
func parseAnalogOut010VVoltage(s string) (float64, bool) {
	v, err := strconv.ParseFloat(s, 64)
	// Written as !(in range) so that NaN, which compares false against
	// everything, is rejected too.
	if err != nil || !(v >= 0 && v <= 10) {
		return 0, false
	}
	return v, true
}

// SetAnalogOut010VChannel sets one output (1-4, as labelled on the board,
// left to right - the repository maps that to the DAC channel) of an
// MCP4728-based 0-10V output module: PUT /api/v1/module/analogout010v/{address}/{channel}/{voltage}
// Optional query parameter persist=true then stores the state of all four
// outputs (not only this one) in the DAC's EEPROM as power-up default - see
// modulesRepository.SetAnalogOut010V. Responds with the full module status as read
// back from the DAC.
func (c ModulesController) SetAnalogOut010VChannel(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !analogOut010VAllowed(w, r, mconf, conf) {
			return
		}

		vars := mux.Vars(r)
		addressStr := vars["address"]

		address, err := parseHexOrDecimal(addressStr)
		if err != nil {
			serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Invalid address: " + err.Error() + ". Must be hex (e.g., 0x60) or decimal (e.g., 96)"})
			return
		}

		channel, err := strconv.Atoi(vars["channel"])
		if err != nil || channel < 1 || channel > 4 {
			serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Invalid channel: must be a number between 1 and 4"})
			return
		}

		voltage, ok := parseAnalogOut010VVoltage(vars["voltage"])
		if !ok {
			serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Voltage must be a number between 0.0 and 10.0 V"})
			return
		}

		persist, _ := strconv.ParseBool(r.URL.Query().Get("persist"))

		moduleRepo := modulesRepository.ModulesRepository{}
		status, err := moduleRepo.SetAnalogOut010V(uint16(address),
			[]modulesRepository.AnalogOut010VSetting{{Channel: channel - 1, Voltage: voltage}}, persist, mconf)
		respondAnalogOut010V(w, addressStr, status, err)
	}
}

// analogOut010VRequest is the JSON body of SetAnalogOut010V. Voltages is
// indexed 0-3 for outputs 1-4 as labelled on the board; a null (or missing trailing) entry leaves
// that channel unchanged.
type analogOut010VRequest struct {
	Voltages []*float64 `json:"voltages"`
	Persist  bool       `json:"persist"`
}

// SetAnalogOut010V sets several channels of an MCP4728-based 0-10V output
// module in one I2C transaction: PUT /api/v1/module/analogout010v/{address}
// with body {"voltages": [1.5, null, 10, 0], "persist": false}. With
// "persist": true the resulting state of all four outputs is stored as
// power-up default; {"voltages": [], "persist": true} stores the current
// state without changing anything.
func (c ModulesController) SetAnalogOut010V(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !analogOut010VAllowed(w, r, mconf, conf) {
			return
		}

		addressStr := mux.Vars(r)["address"]
		address, err := parseHexOrDecimal(addressStr)
		if err != nil {
			serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Invalid address: " + err.Error() + ". Must be hex (e.g., 0x60) or decimal (e.g., 96)"})
			return
		}

		var req analogOut010VRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
			serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Invalid JSON body: " + err.Error()})
			return
		}
		if len(req.Voltages) > 4 {
			serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "voltages must have at most 4 entries (channels 1-4)"})
			return
		}

		var settings []modulesRepository.AnalogOut010VSetting
		for i, v := range req.Voltages {
			if v == nil {
				continue
			}
			if !(*v >= 0 && *v <= 10) {
				serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Voltage of channel " + strconv.Itoa(i+1) + " must be between 0.0 and 10.0 V"})
				return
			}
			settings = append(settings, modulesRepository.AnalogOut010VSetting{Channel: i, Voltage: *v})
		}

		moduleRepo := modulesRepository.ModulesRepository{}
		status, err := moduleRepo.SetAnalogOut010V(uint16(address), settings, req.Persist, mconf)
		respondAnalogOut010V(w, addressStr, status, err)
	}
}

// ReadAnalogOut010V reads back all four channels of an MCP4728-based 0-10V
// output module: GET /api/v1/module/analogout010v/{address}
func (c ModulesController) ReadAnalogOut010V(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !analogOut010VAllowed(w, r, mconf, conf) {
			return
		}

		addressStr := mux.Vars(r)["address"]
		address, err := parseHexOrDecimal(addressStr)
		if err != nil {
			serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Invalid address: " + err.Error() + ". Must be hex (e.g., 0x60) or decimal (e.g., 96)"})
			return
		}

		moduleRepo := modulesRepository.ModulesRepository{}
		status, err := moduleRepo.ReadAnalogOut010V(uint16(address), mconf)
		respondAnalogOut010V(w, addressStr, status, err)
	}
}

// analogOut010VCalibration is the body/response of the calibration
// endpoints. Outputs is indexed 0-3 for outputs 1-4 as labelled on the
// board; each output's measured voltage is gain * nominal + offset (V).
type analogOut010VCalibration struct {
	Moduleaddress string                         `json:"moduleaddress"`
	Calibrated    bool                           `json:"calibrated"`
	Outputs       [4]config.AnalogOutCalibration `json:"outputs"`
}

// AnalogOut010VCalibration reads (GET), stores (PUT, body {"outputs":
// [{"gain":0.9986,"offset":0.0088}, ...4]}) or removes (DELETE) the
// calibration of the 0-10V module at {address} in /etc/smartpiModules:
// /api/v1/module/analogout010v/{address}/calibration
//
// Changing it changes what every output really drives, so the route is
// admin-only (session token of a user in the admin group), not open to
// device tokens with the analogout scope.
func (c ModulesController) AnalogOut010VCalibration(mconf *config.Moduleconfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addressStr := mux.Vars(r)["address"]
		address, err := parseHexOrDecimal(addressStr)
		if err != nil {
			serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Invalid address: " + err.Error() + ". Must be hex (e.g., 0x60) or decimal (e.g., 96)"})
			return
		}

		switch r.Method {
		case http.MethodPut:
			var req analogOut010VCalibration
			if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
				serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "Invalid JSON body: " + err.Error()})
				return
			}
			for i, o := range req.Outputs {
				if err := o.Validate(); err != nil {
					serverutils.RespondWithError(w, http.StatusBadRequest, models.Error{Message: "output " + strconv.Itoa(i+1) + ": " + err.Error()})
					return
				}
			}
			if err := mconf.SaveAnalogOut010VCalibration(uint16(address), &req.Outputs); err != nil {
				serverutils.RespondWithError(w, http.StatusInternalServerError, models.Error{Message: err.Error()})
				return
			}
			log.Infof("analogout010v 0x%02X: calibration set to %s", address, config.FormatAnalogOutCalibration(req.Outputs))
		case http.MethodDelete:
			if err := mconf.SaveAnalogOut010VCalibration(uint16(address), nil); err != nil {
				serverutils.RespondWithError(w, http.StatusInternalServerError, models.Error{Message: err.Error()})
				return
			}
			log.Infof("analogout010v 0x%02X: calibration removed", address)
		}

		cal, ok := mconf.AnalogOut010VCalibrationFor(uint16(address))
		if err := json.NewEncoder(w).Encode(analogOut010VCalibration{Moduleaddress: addressStr, Calibrated: ok, Outputs: cal}); err != nil {
			panic(err)
		}
	}
}
