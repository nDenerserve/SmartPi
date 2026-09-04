package modulescontrollers

import (
	"encoding/json"
	"net/http"

	"github.com/nDenerserve/SmartPi/models"
	config "github.com/nDenerserve/SmartPi/smartpi/config"
	modulesRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/modules"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
)

// ScanI2C reports which addresses are occupied on the configured I2C bus.
//
// Unlike the other module endpoints (digitalout, analogin, analogout420ma)
// this has no per-module username allowlist: it's a read-only diagnostic
// over the whole bus, not an actuator or a per-module reading, so it doesn't
// naturally belong to one module's AllowedXUser list. Access is gated purely
// by TokenVerifyMiddleWare and the i2c:scan scope, the same as the network
// scan endpoint (see controllers.Controller.ScanWifi).
func (c ModulesController) ScanI2C(mconf *config.Moduleconfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var error models.Error

		moduleRepo := modulesRepository.ModulesRepository{}
		status, err := moduleRepo.ScanI2C(mconf)
		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		if err := json.NewEncoder(w).Encode(status); err != nil {
			panic(err)
		}
	}
}
